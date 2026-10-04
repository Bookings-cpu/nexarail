package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/nexarail/chain/x/common"
	"github.com/nexarail/chain/x/settlement/types"
)

// TreasuryModuleAccount is the NexaRail protocol treasury module account name.
// Aliases x/common's canonical constant — see x/common/accounts.go.
const TreasuryModuleAccount = common.TreasuryModuleAccount

// BurnerModuleAccount is the NexaRail burner module account name.
// Aliases x/common's canonical constant — see x/common/accounts.go.
// Requires authtypes.Burner permission.
const BurnerModuleAccount = common.BurnerModuleAccount

// Keeper maintains the state for the settlement module.
type Keeper struct {
	storeKey       storetypes.StoreKey
	authority      string
	merchantKeeper types.MerchantKeeper
	feesKeeper     types.FeesKeeper
	bankKeeper     types.BankKeeper
}

// NewKeeper creates a new settlement keeper.
func NewKeeper(
	storeKey storetypes.StoreKey,
	authority string,
	merchantKeeper types.MerchantKeeper,
	feesKeeper types.FeesKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		storeKey:       storeKey,
		authority:      authority,
		merchantKeeper: merchantKeeper,
		feesKeeper:     feesKeeper,
		bankKeeper:     bankKeeper,
	}
}

func (k Keeper) GetAuthority() string { return k.authority }

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// --- Params ---

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	bz := ctx.KVStore(k.storeKey).Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}
	var p types.Params
	if err := json.Unmarshal(bz, &p); err != nil {
		panic(fmt.Errorf("settlement params unmarshal: %w", err))
	}
	return p
}

func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(p)
	if err != nil {
		return err
	}
	ctx.KVStore(k.storeKey).Set(types.ParamsKey, bz)
	return nil
}

// --- Settlement counter ---

func (k Keeper) getNextSettlementID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SettlementCountKey)
	var count uint64
	if bz != nil {
		count = sdk.BigEndianToUint64(bz)
	}
	count++
	store.Set(types.SettlementCountKey, sdk.Uint64ToBigEndian(count))
	return count
}

// --- Settlement storage ---

func (k Keeper) SetSettlement(ctx sdk.Context, s types.Settlement) error {
	bz, err := json.Marshal(s)
	if err != nil {
		return err
	}
	ctx.KVStore(k.storeKey).Set(types.SettlementKey(s.Id), bz)
	k.SetSettlementIndexes(ctx, s)
	return nil
}

func (k Keeper) GetSettlement(ctx sdk.Context, id uint64) (types.Settlement, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.SettlementKey(id))
	if bz == nil {
		return types.Settlement{}, false
	}
	var s types.Settlement
	if err := json.Unmarshal(bz, &s); err != nil {
		panic(fmt.Errorf("settlement unmarshal id=%d: %w", id, err))
	}
	return s, true
}

func (k Keeper) HasSettlement(ctx sdk.Context, id uint64) bool {
	return ctx.KVStore(k.storeKey).Has(types.SettlementKey(id))
}

func (k Keeper) GetAllSettlements(ctx sdk.Context) []types.Settlement {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.SettlementKeyPrefix)
	defer iter.Close()
	return k.collectSettlements(iter)
}

// --- Indexes ---

func (k Keeper) SetSettlementIndexes(ctx sdk.Context, s types.Settlement) {
	// Index by merchant owner
	store := ctx.KVStore(k.storeKey)
	merchantKey := append([]byte{0x11}, append([]byte(s.MerchantOwner), sdk.Uint64ToBigEndian(s.Id)...)...)
	store.Set(merchantKey, []byte{1})

	// Index by payer
	payerKey := append([]byte{0x12}, append([]byte(s.Payer), sdk.Uint64ToBigEndian(s.Id)...)...)
	store.Set(payerKey, []byte{1})
}

func (k Keeper) GetSettlementsByMerchant(ctx sdk.Context, merchantOwner string) []types.Settlement {
	store := ctx.KVStore(k.storeKey)
	prefix := append([]byte{0x11}, []byte(merchantOwner)...)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var settlements []types.Settlement
	for ; iter.Valid(); iter.Next() {
		// Extract settlement ID from the key
		key := iter.Key()
		if len(key) > len(merchantOwner)+1 {
			id := sdk.BigEndianToUint64(key[len(merchantOwner)+1:])
			if s, ok := k.GetSettlement(ctx, id); ok {
				settlements = append(settlements, s)
			}
		}
	}
	return settlements
}

func (k Keeper) GetSettlementsByPayer(ctx sdk.Context, payer string) []types.Settlement {
	store := ctx.KVStore(k.storeKey)
	prefix := append([]byte{0x12}, []byte(payer)...)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var settlements []types.Settlement
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) > len(payer)+1 {
			id := sdk.BigEndianToUint64(key[len(payer)+1:])
			if s, ok := k.GetSettlement(ctx, id); ok {
				settlements = append(settlements, s)
			}
		}
	}
	return settlements
}

func (k Keeper) RebuildIndexes(ctx sdk.Context) {
	for _, s := range k.GetAllSettlements(ctx) {
		k.SetSettlementIndexes(ctx, s)
	}
}

// --- CreateSettlement ---

func (k Keeper) CreateSettlement(ctx sdk.Context, msg *types.MsgCreateSettlement) (*types.Settlement, error) {
	params := k.GetParams(ctx)

	// 1. Settlements must be enabled
	if !params.Enabled {
		return nil, types.ErrSettlementsDisabled
	}

	// 2. Validate payer
	payerAddr, err := sdk.AccAddressFromBech32(msg.Payer)
	if err != nil {
		return nil, fmt.Errorf("invalid payer: %w", err)
	}

	// 3. Amount must be positive
	if msg.Amount.IsZero() || msg.Amount.IsNegative() {
		return nil, fmt.Errorf("amount: %w", types.ErrAmountNotPositive)
	}

	// 4. Merchant must exist
	merchantAddr, err := sdk.AccAddressFromBech32(msg.MerchantOwner)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant: %w", err)
	}

	if payerAddr.Equals(merchantAddr) {
		return nil, fmt.Errorf("payer cannot settle against their own merchant account: %w", types.ErrInvalidMerchant)
	}

	merchant, found := k.merchantKeeper.GetMerchant(ctx, merchantAddr)
	if !found {
		return nil, fmt.Errorf("merchant %s not found: %w", msg.MerchantOwner, types.ErrInvalidMerchant)
	}

	// 5. Merchant must be active
	if merchant.Status != 0 { // 0 = active in x/merchant
		return nil, fmt.Errorf("merchant %s is not active (status=%d): %w",
			merchant.Owner, merchant.Status, types.ErrMerchantNotActive)
	}

	// 6. Settlement address. Note: there is currently no separate merchant
	// settlement/payout address distinct from the owner address — this was a
	// dead no-op fallback to itself, not an actual alternate-address feature.
	// If merchants need a distinct settlement address later, that needs its
	// own field on the merchant record, not a fallback here.
	settlementAddress := merchant.Owner

	// 7. Generate settlement ID
	id := k.getNextSettlementID(ctx)

	// 8. Fee calculation in basis points.
	// Rounds the fee up (ceiling), not down: with floor division any amount
	// small enough that amount*feeRateBps < 10000 produced baseFee=0 — a
	// structuring loophole letting a payer move funds through the settlement
	// path fee-free by splitting into many sub-threshold settlements. Ceiling
	// division guarantees a nonzero fee for any nonzero amount, and is a
	// no-op whenever amount*feeRateBps already divides evenly (the common
	// case), so it only changes behavior for amounts that were previously
	// under-charged to zero.
	amount := msg.Amount.Amount
	feeRateBps := sdk.NewInt(int64(params.FeeRateBps))
	bpsFactor := sdk.NewInt(10000)

	feeNumerator := amount.Mul(feeRateBps)
	baseFee := feeNumerator.Add(bpsFactor).Sub(sdk.OneInt()).Quo(bpsFactor)

	// 9. Apply merchant rebate tier
	rebateBps := params.GetRebateBps(merchant.RebateTier)
	rebateBpsInt := sdk.NewInt(int64(rebateBps))
	rebateAmount := baseFee.Mul(rebateBpsInt).Quo(bpsFactor)
	netFee := baseFee.Sub(rebateAmount)

	// 10. Read fee split from x/fees
	feeParams := k.feesKeeper.GetParams(ctx)
	valBps := sdk.NewInt(int64(feeParams.ValidatorShareBps))
	treasuryBps := sdk.NewInt(int64(feeParams.TreasuryShareBps))

	valShare := netFee.Mul(valBps).Quo(bpsFactor)
	treasuryShare := netFee.Mul(treasuryBps).Quo(bpsFactor)
	burnShare := netFee.Sub(valShare).Sub(treasuryShare) // remainder → burn

	denom := msg.Amount.Denom
	now := ctx.BlockTime().Unix()

	settlement := types.NewSettlement(
		id,
		payerAddr.String(),
		merchantAddr.String(),
		merchant.Name, // merchant ID = merchant name for now
		settlementAddress,
		sdk.NewCoin(denom, amount),
		sdk.NewCoin(denom, netFee),
		sdk.NewCoin(denom, valShare),
		sdk.NewCoin(denom, treasuryShare),
		sdk.NewCoin(denom, burnShare),
		sdk.NewCoin(denom, rebateAmount),
		rebateBps,
		"", // payment_reference
		"", // memo
		msg.Metadata,
		now,
	)

	// --- Live transfer: payer → merchant (net amount) ---
	fundsSettled := false
	treasuryRouted := false
	burnRouted := false
	if params.LiveEnabled {
		// Calculate merchant net: gross amount minus protocol fee
		merchantNet := amount.Sub(netFee)
		if merchantNet.IsNegative() {
			return nil, fmt.Errorf("merchant net amount is negative: %w", types.ErrAmountNotPositive)
		}

		// Validate merchant settlement address
		merchantSettlementAddr, addrErr := sdk.AccAddressFromBech32(settlementAddress)
		if addrErr != nil {
			return nil, fmt.Errorf("invalid settlement address: %w", addrErr)
		}

		// Transfer merchant net from payer to merchant BEFORE state mutation
		if err := k.bankKeeper.SendCoins(ctx, payerAddr, merchantSettlementAddr,
			sdk.NewCoins(sdk.NewCoin(denom, merchantNet))); err != nil {
			return nil, fmt.Errorf("live settlement transfer failed: %w", err)
		}

		// Transfer validator share from payer to the fee collector module account.
		// The settlement record unconditionally claims this share was collected
		// (ValidatorShare field, set below regardless of routing flags), so unlike
		// treasury/burn routing — which are optional protocol allocations gated by
		// their own flags — this transfer is not optional: it's the base fee
		// collection the record already asserts happened. Previously this was
		// computed and recorded but never actually transferred anywhere, so the
		// payer was silently under-debited by the validator's share of every fee.
		if valShare.IsPositive() {
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, payerAddr,
				feeParams.FeeCollectorName,
				sdk.NewCoins(sdk.NewCoin(denom, valShare))); err != nil {
				return nil, fmt.Errorf("live settlement validator-share routing failed: %w", err)
			}
		}

		// Transfer treasury share from payer to protocol treasury
		if params.TreasuryRoutingEnabled && treasuryShare.IsPositive() {
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, payerAddr,
				TreasuryModuleAccount,
				sdk.NewCoins(sdk.NewCoin(denom, treasuryShare))); err != nil {
				return nil, fmt.Errorf("live settlement treasury routing failed: %w", err)
			}
			treasuryRouted = true
		}

		// Burn burn share via nexarail_burner module account.
		// Requires BurnRoutingEnabled=true AND TreasuryRoutingEnabled=true (and
		// LiveEnabled) — matching the documented contract above, which the
		// original flag check here did not actually enforce.
		if params.BurnRoutingEnabled && params.TreasuryRoutingEnabled && burnShare.IsPositive() {
			// Step 1: Send burn share from payer to burner module account
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, payerAddr,
				BurnerModuleAccount,
				sdk.NewCoins(sdk.NewCoin(denom, burnShare))); err != nil {
				return nil, fmt.Errorf("live settlement burn routing failed: %w", err)
			}
			// Step 2: Burn the coins from the burner module account
			if err := k.bankKeeper.BurnCoins(ctx, BurnerModuleAccount,
				sdk.NewCoins(sdk.NewCoin(denom, burnShare))); err != nil {
				return nil, fmt.Errorf("burn execution failed: %w", err)
			}
			burnRouted = true
		}

		fundsSettled = true
	}

	// Set status to completed
	settlement.FundsSettled = fundsSettled
	settlement.BurnExecuted = burnRouted
	settlement.Status = int32(types.SettlementCompleted)
	settlement.UpdatedAt = now

	if err := k.SetSettlement(ctx, settlement); err != nil {
		return nil, err
	}

	// Emit events
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCreateSettlement,
		sdk.NewAttribute(types.AttributeKeySettlementId, fmt.Sprintf("%d", settlement.Id)),
		sdk.NewAttribute(types.AttributeKeyPayer, settlement.Payer),
		sdk.NewAttribute(types.AttributeKeyMerchantOwner, settlement.MerchantOwner),
		sdk.NewAttribute(types.AttributeKeyAmount, settlement.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyFeeAmount, settlement.FeeAmount.String()),
		sdk.NewAttribute(types.AttributeKeyValidatorShare, settlement.ValidatorShare.String()),
		sdk.NewAttribute(types.AttributeKeyTreasuryShare, settlement.TreasuryShare.String()),
		sdk.NewAttribute(types.AttributeKeyBurnShare, settlement.BurnShare.String()),
		sdk.NewAttribute(types.AttributeKeyRebateBps, fmt.Sprintf("%d", settlement.RebateAppliedBps)),
		sdk.NewAttribute(types.AttributeKeyStatus, fmt.Sprintf("%d", settlement.Status)),
		sdk.NewAttribute(types.AttributeKeyFundsSettled, fmt.Sprintf("%t", settlement.FundsSettled)),
		sdk.NewAttribute(types.AttributeKeyTreasuryRouted, fmt.Sprintf("%t", treasuryRouted)),
		sdk.NewAttribute(types.AttributeKeyBurnRouted, fmt.Sprintf("%t", burnRouted)),
		sdk.NewAttribute(types.AttributeKeyMetadata, settlement.Metadata),
	))

	return &settlement, nil
}

// --- UpdateSettlementStatus (authority-only) ---

func (k Keeper) UpdateSettlementStatus(ctx sdk.Context, authority string, id uint64, newStatus int32) error {
	if authority != k.authority {
		return fmt.Errorf("expected %s, got %s: %w", k.authority, authority, types.ErrUnauthorized)
	}

	s, found := k.GetSettlement(ctx, id)
	if !found {
		return fmt.Errorf("settlement %d: %w", id, types.ErrSettlementNotFound)
	}

	// Terminal statuses cannot transition
	if s.IsTerminal() {
		return fmt.Errorf("settlement %d is in terminal state %s: %w",
			id, types.SettlementStatus(s.Status), types.ErrInvalidStatusTransition)
	}

	// Live-settled records (FundsSettled=true) cannot change status
	// because no automated refund mechanism exists yet (Phase 5F.2)
	if s.FundsSettled {
		return fmt.Errorf("settlement %d is live-settled (funds_settled=true): status changes blocked in Phase 5F.2", id)
	}

	// New status must be valid (0-4)
	if newStatus < 0 || newStatus > 4 {
		return fmt.Errorf("status %d: %w", newStatus, types.ErrInvalidStatus)
	}

	// Cannot transition back to pending from completed
	if s.Status == int32(types.SettlementCompleted) && newStatus == int32(types.SettlementPending) {
		return fmt.Errorf("cannot transition completed back to pending: %w", types.ErrInvalidStatusTransition)
	}

	s.Status = newStatus
	s.UpdatedAt = ctx.BlockTime().Unix()

	if err := k.SetSettlement(ctx, s); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateSettlementStatus,
		sdk.NewAttribute(types.AttributeKeySettlementId, fmt.Sprintf("%d", s.Id)),
		sdk.NewAttribute(types.AttributeKeyStatus, fmt.Sprintf("%d", s.Status)),
		sdk.NewAttribute(types.AttributeKeyAuthority, authority),
	))

	return nil
}

// --- UpdateParams (authority-only) ---

func (k Keeper) UpdateParams(ctx sdk.Context, authority string, params types.Params) error {
	if authority != k.authority {
		return fmt.Errorf("expected %s, got %s: %w", k.authority, authority, types.ErrUnauthorized)
	}
	return k.SetParams(ctx, params)
}

// --- invariant helpers ---

// ActiveSettledTotals returns the sum of merchant_net across all live-settled records
// (FundsSettled=true, status=Completed). This represents total settlement volume on-chain.
func (k Keeper) ActiveSettledTotals(ctx sdk.Context) sdk.Coins {
	totals := sdk.Coins{}
	for _, s := range k.GetAllSettlements(ctx) {
		if s.FundsSettled && s.Status == int32(types.SettlementCompleted) {
			merchantNet := s.Amount.Amount.Sub(s.FeeAmount.Amount)
			totals = totals.Add(sdk.NewCoin(s.Amount.Denom, merchantNet))
		}
	}
	return totals
}

// ValidateSettlementFundsInvariant checks consistency of FundsSettled and BurnExecuted against status.
func (k Keeper) ValidateSettlementFundsInvariant(ctx sdk.Context) error {
	for _, s := range k.GetAllSettlements(ctx) {
		if s.FundsSettled {
			if s.Status != int32(types.SettlementCompleted) {
				return fmt.Errorf(
					"invariant violation: settlement %d has FundsSettled=true but status=%s",
					s.Id, types.SettlementStatus(s.Status),
				)
			}
			merchantNet := s.Amount.Amount.Sub(s.FeeAmount.Amount)
			if !merchantNet.IsPositive() {
				return fmt.Errorf(
					"invariant violation: settlement %d has FundsSettled=true but merchant net %s is not positive",
					s.Id, merchantNet,
				)
			}
			netPlusFee := s.FeeAmount.Amount.Add(merchantNet)
			if netPlusFee.GT(s.Amount.Amount) {
				return fmt.Errorf(
					"invariant violation: settlement %d net+merchant %s > gross %s",
					s.Id, netPlusFee, s.Amount.Amount,
				)
			}
		}
		if s.BurnExecuted {
			if !s.FundsSettled {
				return fmt.Errorf(
					"invariant violation: settlement %d has BurnExecuted=true but FundsSettled=false",
					s.Id,
				)
			}
			if s.Status != int32(types.SettlementCompleted) {
				return fmt.Errorf(
					"invariant violation: settlement %d has BurnExecuted=true but status=%s",
					s.Id, types.SettlementStatus(s.Status),
				)
			}
			if !s.BurnShare.Amount.IsPositive() {
				return fmt.Errorf(
					"invariant violation: settlement %d has BurnExecuted=true but burn share %s is not positive",
					s.Id, s.BurnShare,
				)
			}
		}
	}
	return nil
}

// --- helpers ---

func (k Keeper) collectSettlements(iter sdk.Iterator) []types.Settlement {
	var settlements []types.Settlement
	for ; iter.Valid(); iter.Next() {
		var s types.Settlement
		if err := json.Unmarshal(iter.Value(), &s); err != nil {
			continue
		}
		settlements = append(settlements, s)
	}
	return settlements
}
