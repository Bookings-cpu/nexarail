package keeper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/nexarail/chain/x/escrow/types"
)

type Keeper struct {
	storeKey       storetypes.StoreKey
	authority      string
	merchantKeeper types.MerchantKeeper
	bankKeeper     types.BankKeeper
}

func NewKeeper(storeKey storetypes.StoreKey, authority string, mk types.MerchantKeeper, bk types.BankKeeper) Keeper {
	return Keeper{storeKey: storeKey, authority: authority, merchantKeeper: mk, bankKeeper: bk}
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
		panic(fmt.Errorf("escrow params: %w", err))
	}
	return p
}

func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}
	bz, _ := json.Marshal(p)
	ctx.KVStore(k.storeKey).Set(types.ParamsKey, bz)
	return nil
}

func (k Keeper) UpdateParams(ctx sdk.Context, authority string, p types.Params) error {
	if authority != k.authority {
		return fmt.Errorf("%w", types.ErrUnauthorized)
	}
	return k.SetParams(ctx, p)
}

// --- Escrow storage ---

func (k Keeper) SetEscrow(ctx sdk.Context, e types.Escrow) error {
	bz, err := json.Marshal(e)
	if err != nil {
		return err
	}
	ctx.KVStore(k.storeKey).Set(types.EscrowKey(e.EscrowId), bz)
	k.SetEscrowIndexes(ctx, e)
	return nil
}

func (k Keeper) GetEscrow(ctx sdk.Context, id string) (types.Escrow, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.EscrowKey(id))
	if bz == nil {
		return types.Escrow{}, false
	}
	var e types.Escrow
	if err := json.Unmarshal(bz, &e); err != nil {
		panic(fmt.Errorf("escrow %s: %w", id, err))
	}
	return e, true
}

func (k Keeper) HasEscrow(ctx sdk.Context, id string) bool {
	return ctx.KVStore(k.storeKey).Has(types.EscrowKey(id))
}

func (k Keeper) GetAllEscrows(ctx sdk.Context) []types.Escrow {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.EscrowKeyPrefix)
	defer iter.Close()
	var escrows []types.Escrow
	for ; iter.Valid(); iter.Next() {
		var e types.Escrow
		if err := json.Unmarshal(iter.Value(), &e); err != nil {
			continue
		}
		escrows = append(escrows, e)
	}
	return escrows
}

// --- Indexes ---

func (k Keeper) SetEscrowIndexes(ctx sdk.Context, e types.Escrow) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.EscrowByBuyerKey(e.BuyerAddress, e.EscrowId), []byte{1})
	store.Set(types.EscrowBySellerKey(e.SellerAddress, e.EscrowId), []byte{1})
	store.Set(types.EscrowByMerchantKey(e.MerchantId, e.EscrowId), []byte{1})
}

func (k Keeper) GetEscrowsByBuyer(ctx sdk.Context, buyer string) []types.Escrow {
	return k.escrowsByPrefix(ctx, types.EscrowByBuyerPrefix, []byte(buyer))
}

func (k Keeper) GetEscrowsBySeller(ctx sdk.Context, seller string) []types.Escrow {
	return k.escrowsByPrefix(ctx, types.EscrowBySellerPrefix, []byte(seller))
}

func (k Keeper) GetEscrowsByMerchant(ctx sdk.Context, merchantId string) []types.Escrow {
	return k.escrowsByPrefix(ctx, types.EscrowByMerchantPrefix, []byte(merchantId))
}

func (k Keeper) escrowsByPrefix(ctx sdk.Context, prefix, key []byte) []types.Escrow {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, append(prefix, key...))
	defer iter.Close()
	var escrows []types.Escrow
	for ; iter.Valid(); iter.Next() {
		id := string(iter.Key()[len(prefix)+len(key):])
		if e, ok := k.GetEscrow(ctx, id); ok {
			escrows = append(escrows, e)
		}
	}
	return escrows
}

func (k Keeper) RebuildIndexes(ctx sdk.Context) {
	for _, e := range k.GetAllEscrows(ctx) {
		k.SetEscrowIndexes(ctx, e)
	}
}

// ActiveCustodiedEscrowTotals returns the total coins held in active custodied escrows per denom.
func (k Keeper) ActiveCustodiedEscrowTotals(ctx sdk.Context) sdk.Coins {
	totals := sdk.Coins{}
	for _, e := range k.GetAllEscrows(ctx) {
		if e.FundsCustodied {
			totals = totals.Add(e.Amount)
		}
	}
	return totals
}

// ValidateCustodyInvariant checks that no terminal escrow has FundsCustodied=true.
func (k Keeper) ValidateCustodyInvariant(ctx sdk.Context) error {
	for _, e := range k.GetAllEscrows(ctx) {
		if e.FundsCustodied {
			if e.Status == int32(types.EscrowReleased) || e.Status == int32(types.EscrowRefunded) || e.Status == int32(types.EscrowCancelled) {
				return fmt.Errorf("invariant violation: escrow %s is terminal (%s) but funds_custodied=true", e.EscrowId, types.EscrowStatus(e.Status))
			}
		}
	}
	return nil
}

// --- CreateEscrow ---

func (k Keeper) CreateEscrow(ctx sdk.Context, msg *types.MsgCreateEscrow) (*types.Escrow, error) {
	params := k.GetParams(ctx)
	if !params.EscrowsEnabled {
		return nil, types.ErrEscrowsDisabled
	}
	if k.HasEscrow(ctx, msg.EscrowId) {
		return nil, fmt.Errorf("%s: %w", msg.EscrowId, types.ErrEscrowExists)
	}

	// Validate buyer
	buyer, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return nil, fmt.Errorf("buyer: %w", err)
	}

	// Validate seller
	seller, err := sdk.AccAddressFromBech32(msg.SellerAddress)
	if err != nil {
		return nil, fmt.Errorf("seller: %w", err)
	}
	if msg.Buyer == msg.SellerAddress {
		return nil, fmt.Errorf("%w", types.ErrInvalidBuyer)
	}

	// Validate merchant
	merchantID := strings.TrimSpace(msg.MerchantId)
	if merchantID == "" {
		return nil, fmt.Errorf("%w", types.ErrInvalidMerchantID)
	}
	merchant, found := k.merchantKeeper.GetMerchant(ctx, seller)
	if !found {
		return nil, fmt.Errorf("merchant %s: %w", msg.SellerAddress, types.ErrInvalidMerchantID)
	}
	if merchant.Status != 0 {
		return nil, fmt.Errorf("merchant %s status=%d: %w", merchant.Owner, merchant.Status, types.ErrMerchantNotActive)
	}

	// Expiry
	now := ctx.BlockTime().Unix()
	expires := msg.ExpiresAt
	if expires == 0 {
		expires = now + int64(params.DefaultExpirySeconds)
	}

	e := types.NewEscrow(msg.EscrowId, buyer.String(), seller.String(), merchantID, msg.AssetDenom, msg.Amount,
		msg.PaymentReference, msg.Memo, now, expires)
	e.Status = int32(types.EscrowCreated)
	e.DisputeStatus = int32(types.DisputeNone)

	// Full structural + amount validation, including the AssetDenom == Amount.Denom
	// consistency check and the correctly-ordered minimum-amount check. This was
	// previously only called from genesis import, not here — the inline checks
	// that used to live in this function keyed the minimum-amount check on the
	// wrong field (AssetDenom instead of Amount.Denom) and could panic on a
	// denom-mismatched Amount, since sdk.Coin.IsLT panics across denoms.
	if err := e.ValidateWithParams(params); err != nil {
		return nil, err
	}

	// Live custody: transfer buyer → escrow module account
	if params.LiveEnabled {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, buyer, types.EscrowModuleAccount, sdk.NewCoins(msg.Amount)); err != nil {
			return nil, fmt.Errorf("live escrow funding failed: %w", err)
		}
		e.Status = int32(types.EscrowFunded)
		e.FundsCustodied = true
	}

	if err := k.SetEscrow(ctx, e); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCreateEscrow,
		sdk.NewAttribute(types.AttrEscrowId, e.EscrowId),
		sdk.NewAttribute(types.AttrBuyer, e.BuyerAddress),
		sdk.NewAttribute(types.AttrSeller, e.SellerAddress),
		sdk.NewAttribute(types.AttrAmount, e.Amount.String()),
		sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", e.Status)),
	))
	return &e, nil
}

// --- ReleaseEscrow ---

func (k Keeper) ReleaseEscrow(ctx sdk.Context, msg *types.MsgReleaseEscrow) error {
	e, found := k.GetEscrow(ctx, msg.EscrowId)
	if !found {
		return fmt.Errorf("%s: %w", msg.EscrowId, types.ErrEscrowNotFound)
	}

	// Signer must be buyer or authority
	if msg.Signer != e.BuyerAddress && msg.Signer != k.authority {
		return fmt.Errorf("%w", types.ErrUnauthorized)
	}

	// Must be CREATED or FUNDED — this already excludes EscrowDisputed (and every
	// other status), so no separate disputed-status check is needed here.
	if e.Status != int32(types.EscrowCreated) && e.Status != int32(types.EscrowFunded) {
		return fmt.Errorf("status %s: %w", types.EscrowStatus(e.Status), types.ErrInvalidTransition)
	}

	// Live custody: transfer escrow module → seller
	if e.FundsCustodied {
		seller, err := sdk.AccAddressFromBech32(e.SellerAddress)
		if err != nil {
			return fmt.Errorf("stored seller_address %q unparseable: %w", e.SellerAddress, err)
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.EscrowModuleAccount, seller, sdk.NewCoins(e.Amount)); err != nil {
			return fmt.Errorf("live release failed: %w", err)
		}
		e.FundsCustodied = false
	}

	e.Status = int32(types.EscrowReleased)
	e.ReleaseReference = strings.TrimSpace(msg.ReleaseReference)
	if msg.Memo != "" {
		e.Memo = strings.TrimSpace(msg.Memo)
	}
	e.UpdatedAt = ctx.BlockTime().Unix()

	if err := k.SetEscrow(ctx, e); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventReleaseEscrow,
		sdk.NewAttribute(types.AttrEscrowId, e.EscrowId),
		sdk.NewAttribute(types.AttrSigner, msg.Signer),
		sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", e.Status)),
	))
	return nil
}

// --- RefundEscrow ---

func (k Keeper) RefundEscrow(ctx sdk.Context, msg *types.MsgRefundEscrow) error {
	e, found := k.GetEscrow(ctx, msg.EscrowId)
	if !found {
		return fmt.Errorf("%s: %w", msg.EscrowId, types.ErrEscrowNotFound)
	}

	if msg.Signer != e.SellerAddress && msg.Signer != k.authority {
		return fmt.Errorf("%w", types.ErrUnauthorized)
	}

	if e.Status != int32(types.EscrowCreated) && e.Status != int32(types.EscrowFunded) {
		return fmt.Errorf("status %s: %w", types.EscrowStatus(e.Status), types.ErrInvalidTransition)
	}

	// Live custody: transfer escrow module → buyer
	if e.FundsCustodied {
		buyer, err := sdk.AccAddressFromBech32(e.BuyerAddress)
		if err != nil {
			return fmt.Errorf("stored buyer_address %q unparseable: %w", e.BuyerAddress, err)
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.EscrowModuleAccount, buyer, sdk.NewCoins(e.Amount)); err != nil {
			return fmt.Errorf("live refund failed: %w", err)
		}
		e.FundsCustodied = false
	}

	e.Status = int32(types.EscrowRefunded)
	e.RefundReference = strings.TrimSpace(msg.RefundReference)
	if msg.Memo != "" {
		e.Memo = strings.TrimSpace(msg.Memo)
	}
	e.UpdatedAt = ctx.BlockTime().Unix()

	if err := k.SetEscrow(ctx, e); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventRefundEscrow,
		sdk.NewAttribute(types.AttrEscrowId, e.EscrowId),
		sdk.NewAttribute(types.AttrSigner, msg.Signer),
		sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", e.Status)),
	))
	return nil
}

// --- OpenDispute ---

func (k Keeper) OpenDispute(ctx sdk.Context, msg *types.MsgOpenDispute) error {
	e, found := k.GetEscrow(ctx, msg.EscrowId)
	if !found {
		return fmt.Errorf("%s: %w", msg.EscrowId, types.ErrEscrowNotFound)
	}

	// Signer must be buyer or seller
	if msg.Signer != e.BuyerAddress && msg.Signer != e.SellerAddress {
		return fmt.Errorf("only buyer/seller can dispute: %w", types.ErrUnauthorized)
	}

	if e.Status != int32(types.EscrowCreated) && e.Status != int32(types.EscrowFunded) {
		return fmt.Errorf("status %s: %w", types.EscrowStatus(e.Status), types.ErrInvalidTransition)
	}

	e.Status = int32(types.EscrowDisputed)
	e.DisputeStatus = int32(types.DisputeOpen)
	e.DisputeReason = strings.TrimSpace(msg.DisputeReason)
	e.UpdatedAt = ctx.BlockTime().Unix()

	if err := k.SetEscrow(ctx, e); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventDisputeEscrow,
		sdk.NewAttribute(types.AttrEscrowId, e.EscrowId),
		sdk.NewAttribute(types.AttrSigner, msg.Signer),
		sdk.NewAttribute(types.AttrReason, e.DisputeReason),
		sdk.NewAttribute(types.AttrDisputeStatus, fmt.Sprintf("%d", e.DisputeStatus)),
	))
	return nil
}

// --- ResolveDispute ---

func (k Keeper) ResolveDispute(ctx sdk.Context, msg *types.MsgResolveDispute) error {
	if msg.Authority != k.authority {
		return fmt.Errorf("%w", types.ErrUnauthorized)
	}

	e, found := k.GetEscrow(ctx, msg.EscrowId)
	if !found {
		return fmt.Errorf("%s: %w", msg.EscrowId, types.ErrEscrowNotFound)
	}

	if e.Status != int32(types.EscrowDisputed) {
		return fmt.Errorf("status %s: %w", types.EscrowStatus(e.Status), types.ErrInvalidTransition)
	}

	ds := msg.DisputeStatus

	// Live custody: resolve based on dispute outcome
	if e.FundsCustodied {
		buyer, err := sdk.AccAddressFromBech32(e.BuyerAddress)
		if err != nil {
			return fmt.Errorf("stored buyer_address %q unparseable: %w", e.BuyerAddress, err)
		}
		seller, err := sdk.AccAddressFromBech32(e.SellerAddress)
		if err != nil {
			return fmt.Errorf("stored seller_address %q unparseable: %w", e.SellerAddress, err)
		}
		switch ds {
		case int32(types.DisputeBuyerWins):
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.EscrowModuleAccount, buyer, sdk.NewCoins(e.Amount)); err != nil {
				return fmt.Errorf("dispute refund failed: %w", err)
			}
			e.FundsCustodied = false
			e.Status = int32(types.EscrowRefunded)
		case int32(types.DisputeSellerWins), int32(types.DisputeSettled):
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.EscrowModuleAccount, seller, sdk.NewCoins(e.Amount)); err != nil {
				return fmt.Errorf("dispute release failed: %w", err)
			}
			e.FundsCustodied = false
			e.Status = int32(types.EscrowReleased)
		case int32(types.DisputeRejected):
			e.Status = int32(types.EscrowCreated)
		default:
			return fmt.Errorf("dispute_status %d: %w", ds, types.ErrInvalidDisputeStatus)
		}
	} else {
		switch ds {
		case int32(types.DisputeBuyerWins):
			e.Status = int32(types.EscrowRefunded)
		case int32(types.DisputeSellerWins), int32(types.DisputeSettled):
			e.Status = int32(types.EscrowReleased)
		case int32(types.DisputeRejected):
			e.Status = int32(types.EscrowCreated)
		default:
			return fmt.Errorf("dispute_status %d: %w", ds, types.ErrInvalidDisputeStatus)
		}
	}

	e.DisputeStatus = ds
	e.ResolutionNote = strings.TrimSpace(msg.ResolutionNote)
	e.UpdatedAt = ctx.BlockTime().Unix()

	if err := k.SetEscrow(ctx, e); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventResolveDispute,
		sdk.NewAttribute(types.AttrEscrowId, e.EscrowId),
		sdk.NewAttribute(types.AttrDisputeStatus, fmt.Sprintf("%d", e.DisputeStatus)),
		sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", e.Status)),
		sdk.NewAttribute(types.AttrResolutionNote, e.ResolutionNote),
	))
	return nil
}

// --- CancelEscrow ---

func (k Keeper) CancelEscrow(ctx sdk.Context, msg *types.MsgCancelEscrow) error {
	e, found := k.GetEscrow(ctx, msg.EscrowId)
	if !found {
		return fmt.Errorf("%s: %w", msg.EscrowId, types.ErrEscrowNotFound)
	}

	if msg.Signer != e.BuyerAddress && msg.Signer != k.authority {
		return fmt.Errorf("%w", types.ErrUnauthorized)
	}

	// The buyer may only cancel before the escrow is funded — this is the
	// "cancel my order before paying" path. Once funds are in custody, the
	// module already provides the correct exit paths: ReleaseEscrow (buyer
	// confirms, pays seller), RefundEscrow (seller voluntarily returns funds
	// to buyer), or OpenDispute/ResolveDispute (arbitrated by governance).
	// Allowing unilateral buyer cancellation on a FUNDED escrow bypassed all
	// of that and let the buyer reclaim 100% of custodied funds at any time
	// with no seller recourse, which defeats the purpose of the module.
	// Authority (governance) retains the ability to cancel in either status
	// for legitimate administrative cleanup.
	if msg.Signer == e.BuyerAddress && msg.Signer != k.authority {
		if e.Status != int32(types.EscrowCreated) {
			return fmt.Errorf("escrow is funded — use release, refund, or dispute instead of cancel: %w", types.ErrInvalidTransition)
		}
	} else if e.Status != int32(types.EscrowCreated) && e.Status != int32(types.EscrowFunded) {
		return fmt.Errorf("only CREATED or FUNDED escrows can be cancelled, got %s: %w",
			types.EscrowStatus(e.Status), types.ErrInvalidTransition)
	}

	// Live custody: transfer escrow module → buyer
	if e.FundsCustodied {
		buyer, err := sdk.AccAddressFromBech32(e.BuyerAddress)
		if err != nil {
			return fmt.Errorf("stored buyer_address %q unparseable: %w", e.BuyerAddress, err)
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.EscrowModuleAccount, buyer, sdk.NewCoins(e.Amount)); err != nil {
			return fmt.Errorf("live cancel failed: %w", err)
		}
		e.FundsCustodied = false
	}

	e.Status = int32(types.EscrowCancelled)
	if msg.Memo != "" {
		e.Memo = strings.TrimSpace(msg.Memo)
	}
	e.UpdatedAt = ctx.BlockTime().Unix()

	if err := k.SetEscrow(ctx, e); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCancelEscrow,
		sdk.NewAttribute(types.AttrEscrowId, e.EscrowId),
		sdk.NewAttribute(types.AttrSigner, msg.Signer),
		sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", e.Status)),
	))
	return nil
}
