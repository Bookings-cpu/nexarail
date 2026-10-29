package app

import (
	"encoding/json"
	"log"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// ExportAppStateAndValidators exports the state of the application for a restart.
func (app *NexaRailApp) ExportAppStateAndValidators(
	forZeroHeight bool,
	jailAllowedAddrs []string,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	// Use the checkState (isCheckTx=true); deliverState is nil outside a block,
	// which is what caused the previous nil-pointer panic here.
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})

	height := app.LastBlockHeight() + 1
	if forZeroHeight {
		height = 0
		app.prepForZeroHeightGenesis(ctx, jailAllowedAddrs)
	}

	genState := app.mm.ExportGenesis(ctx, app.appCodec)
	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := staking.WriteValidators(ctx, app.StakingKeeper)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          height,
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, nil
}

// prepForZeroHeightGenesis prepares the application for a zero-height genesis restart.
//
// Adapted from the canonical cosmos-sdk simapp implementation, with one
// deliberate deviation for NexaRail's coordinated halt recovery:
//
//	jailAllowedAddrs is treated as the *exclusive* set of validators that should
//	remain in the active set. Addresses in the list are un-jailed and re-bonded;
//	every other validator is jailed and removed from the power index. This lets a
//	restart genesis pin the active set to a known-online subset (the coordinators)
//	so the relaunched chain can reach 2/3 without the offline validators.
func (app *NexaRailApp) prepForZeroHeightGenesis(ctx sdk.Context, jailAllowedAddrs []string) {
	applyAllowedAddrs := len(jailAllowedAddrs) > 0

	allowedAddrsMap := make(map[string]bool)
	for _, addr := range jailAllowedAddrs {
		if _, err := sdk.ValAddressFromBech32(addr); err != nil {
			log.Fatal(err)
		}
		allowedAddrsMap[addr] = true
	}

	/* Handle fee distribution state. */

	// withdraw all validator commission
	app.StakingKeeper.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
		_, _ = app.DistrKeeper.WithdrawValidatorCommission(ctx, val.GetOperator())
		return false
	})

	// withdraw all delegator rewards
	dels := app.StakingKeeper.GetAllDelegations(ctx)
	for _, delegation := range dels {
		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		delAddr := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)
		_, _ = app.DistrKeeper.WithdrawDelegationRewards(ctx, delAddr, valAddr)
	}

	// clear validator slash events
	app.DistrKeeper.DeleteAllValidatorSlashEvents(ctx)

	// clear validator historical rewards
	app.DistrKeeper.DeleteAllValidatorHistoricalRewards(ctx)

	// set context height to zero
	height := ctx.BlockHeight()
	ctx = ctx.WithBlockHeight(0)

	// reinitialize all validators
	app.StakingKeeper.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
		// donate any unwithdrawn outstanding reward fraction tokens to the community pool
		scraps := app.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, val.GetOperator())
		feePool := app.DistrKeeper.GetFeePool(ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(scraps...)
		app.DistrKeeper.SetFeePool(ctx, feePool)

		if err := app.DistrKeeper.Hooks().AfterValidatorCreated(ctx, val.GetOperator()); err != nil {
			panic(err)
		}
		return false
	})

	// reinitialize all delegations
	for _, del := range dels {
		valAddr, err := sdk.ValAddressFromBech32(del.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		delAddr := sdk.MustAccAddressFromBech32(del.DelegatorAddress)

		if err := app.DistrKeeper.Hooks().BeforeDelegationCreated(ctx, delAddr, valAddr); err != nil {
			panic(err)
		}
		if err := app.DistrKeeper.Hooks().AfterDelegationModified(ctx, delAddr, valAddr); err != nil {
			panic(err)
		}
	}

	// reset context height
	ctx = ctx.WithBlockHeight(height)

	/* Handle staking state. */

	// iterate through redelegations, reset creation height
	app.StakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingtypes.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		app.StakingKeeper.SetRedelegation(ctx, red)
		return false
	})

	// iterate through unbonding delegations, reset creation height
	app.StakingKeeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd stakingtypes.UnbondingDelegation) (stop bool) {
		for i := range ubd.Entries {
			ubd.Entries[i].CreationHeight = 0
		}
		app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
		return false
	})

	// Iterate through validators by power descending, reset bond heights, apply
	// the allow-list (un-jail coordinators, jail everyone else), and fix the
	// power index accordingly.
	store := ctx.KVStore(app.keys[stakingtypes.StoreKey])
	iter := sdk.KVStoreReversePrefixIterator(store, stakingtypes.ValidatorsKey)
	counter := int16(0)

	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(stakingtypes.AddressFromValidatorsKey(iter.Key()))
		validator, found := app.StakingKeeper.GetValidator(ctx, addr)
		if !found {
			panic("expected validator, not found")
		}

		validator.UnbondingHeight = 0

		if applyAllowedAddrs {
			if allowedAddrsMap[addr.String()] {
				// coordinator: un-jail and ensure it is in the power index so
				// ApplyAndReturnValidatorSetUpdates can bond it.
				validator.Jailed = false
				app.StakingKeeper.SetValidator(ctx, validator)
				app.StakingKeeper.SetValidatorByPowerIndex(ctx, validator)
			} else {
				// non-coordinator: jail and remove from the power index so it
				// cannot be selected into the active set.
				validator.Jailed = true
				app.StakingKeeper.SetValidator(ctx, validator)
				app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validator)
			}
		} else {
			app.StakingKeeper.SetValidator(ctx, validator)
		}

		counter++
	}

	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}

	if _, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx); err != nil {
		log.Fatal(err)
	}

	/* Handle slashing state. */

	// reset start height on signing infos
	app.SlashingKeeper.IterateValidatorSigningInfos(
		ctx,
		func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
			info.StartHeight = 0
			app.SlashingKeeper.SetValidatorSigningInfo(ctx, addr, info)
			return false
		},
	)
}
