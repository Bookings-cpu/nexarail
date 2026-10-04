package keeper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/nexarail/chain/x/payout/types"
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
		panic(err)
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
		return types.ErrUnauthorized
	}
	return k.SetParams(ctx, p)
}

// --- Payout storage ---
func (k Keeper) SetPayout(ctx sdk.Context, p types.Payout) error {
	bz, err := json.Marshal(p)
	if err != nil {
		return err
	}
	ctx.KVStore(k.storeKey).Set(types.PayoutKey(p.PayoutId), bz)
	k.SetPayoutIndexes(ctx, p)
	return nil
}
func (k Keeper) GetPayout(ctx sdk.Context, id string) (types.Payout, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.PayoutKey(id))
	if bz == nil {
		return types.Payout{}, false
	}
	var p types.Payout
	if err := json.Unmarshal(bz, &p); err != nil {
		panic(err)
	}
	return p, true
}
func (k Keeper) HasPayout(ctx sdk.Context, id string) bool {
	return ctx.KVStore(k.storeKey).Has(types.PayoutKey(id))
}
func (k Keeper) GetAllPayouts(ctx sdk.Context) []types.Payout {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PayoutKeyPrefix)
	defer iter.Close()
	var ps []types.Payout
	for ; iter.Valid(); iter.Next() {
		var p types.Payout
		if json.Unmarshal(iter.Value(), &p) == nil {
			ps = append(ps, p)
		}
	}
	return ps
}

// --- Batch storage ---
func (k Keeper) SetBatchPayout(ctx sdk.Context, b types.BatchPayout) error {
	bz, _ := json.Marshal(b)
	ctx.KVStore(k.storeKey).Set(types.BatchPayoutKey(b.BatchId), bz)
	return nil
}
func (k Keeper) GetBatchPayout(ctx sdk.Context, id string) (types.BatchPayout, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.BatchPayoutKey(id))
	if bz == nil {
		return types.BatchPayout{}, false
	}
	var b types.BatchPayout
	json.Unmarshal(bz, &b)
	return b, true
}
func (k Keeper) HasBatchPayout(ctx sdk.Context, id string) bool {
	return ctx.KVStore(k.storeKey).Has(types.BatchPayoutKey(id))
}
func (k Keeper) GetAllBatchPayouts(ctx sdk.Context) []types.BatchPayout {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.BatchPayoutKeyPrefix)
	defer iter.Close()
	var bs []types.BatchPayout
	for ; iter.Valid(); iter.Next() {
		var b types.BatchPayout
		if json.Unmarshal(iter.Value(), &b) == nil {
			bs = append(bs, b)
		}
	}
	return bs
}

// --- Indexes ---
func (k Keeper) SetPayoutIndexes(ctx sdk.Context, p types.Payout) {
	s := ctx.KVStore(k.storeKey)
	s.Set(types.PayoutByMerchantKey(p.MerchantId, p.PayoutId), []byte{1})
	s.Set(types.PayoutByRecipientKey(p.RecipientAddress, p.PayoutId), []byte{1})
	s.Set(types.PayoutByInitiatorKey(p.InitiatorAddress, p.PayoutId), []byte{1})
}
func (k Keeper) RebuildIndexes(ctx sdk.Context) {
	for _, p := range k.GetAllPayouts(ctx) {
		k.SetPayoutIndexes(ctx, p)
	}
}

func (k Keeper) GetPayoutsByMerchant(ctx sdk.Context, m string) []types.Payout {
	return k.byPrefix(ctx, []byte{0x11}, []byte(m))
}
func (k Keeper) GetPayoutsByRecipient(ctx sdk.Context, r string) []types.Payout {
	return k.byPrefix(ctx, []byte{0x12}, []byte(r))
}
func (k Keeper) GetPayoutsByInitiator(ctx sdk.Context, i string) []types.Payout {
	return k.byPrefix(ctx, []byte{0x13}, []byte(i))
}
func (k Keeper) byPrefix(ctx sdk.Context, prefix, key []byte) []types.Payout {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), append(prefix, key...))
	defer iter.Close()
	var ps []types.Payout
	for ; iter.Valid(); iter.Next() {
		id := string(iter.Key()[len(prefix)+len(key):])
		if p, ok := k.GetPayout(ctx, id); ok {
			ps = append(ps, p)
		}
	}
	return ps
}

// --- validateMerchant ---
func (k Keeper) validateMerchant(ctx sdk.Context, addr, merchantID string) error {
	if strings.TrimSpace(merchantID) == "" {
		return types.ErrInvalidMerchantID
	}
	owner, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return fmt.Errorf("merchant addr: %w", err)
	}
	m, found := k.merchantKeeper.GetMerchant(ctx, owner)
	if !found {
		return fmt.Errorf("merchant %s: %w", addr, types.ErrInvalidMerchantID)
	}
	if m.Status != 0 {
		return fmt.Errorf("merchant %s status=%d: %w", m.Owner, m.Status, types.ErrMerchantNotActive)
	}
	return nil
}

// --- CreatePayout ---
func (k Keeper) CreatePayout(ctx sdk.Context, msg *types.MsgCreatePayout) error {
	params := k.GetParams(ctx)
	if !params.PayoutsEnabled {
		return types.ErrPayoutsDisabled
	}
	if k.HasPayout(ctx, msg.PayoutId) {
		return fmt.Errorf("%s: %w", msg.PayoutId, types.ErrPayoutExists)
	}
	if err := k.validateMerchant(ctx, msg.RecipientAddress, msg.MerchantId); err != nil {
		return err
	}
	if msg.Initiator == msg.RecipientAddress {
		return types.ErrInvalidInitiator
	}
	if msg.Amount.IsZero() || msg.Amount.IsNegative() {
		return types.ErrAmountNotPositive
	}
	if msg.Amount.Denom == params.MinPayoutAmount.Denom && msg.Amount.IsLT(params.MinPayoutAmount) {
		return fmt.Errorf("below min: %w", types.ErrAmountNotPositive)
	}

	now := ctx.BlockTime().Unix()
	p := types.NewPayout(msg.PayoutId, "", msg.MerchantId, msg.Initiator, msg.RecipientAddress, msg.AssetDenom, msg.Amount, msg.PayoutType, msg.PayoutReference, msg.Memo, now)
	if params.ApprovalRequired {
		p.Status = int32(types.PayoutCreated)
	} else {
		p.Status = int32(types.PayoutApproved)
		p.ApprovedAt = now
	}

	// Full structural validation (PayoutId format, amount/denom consistency,
	// reference/memo length caps) as a final check before this enters state.
	if err := p.ValidateWithParams(params); err != nil {
		return err
	}

	if err := k.SetPayout(ctx, p); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCreatePayout,
		sdk.NewAttribute(types.AttrPayoutId, p.PayoutId), sdk.NewAttribute(types.AttrInitiator, p.InitiatorAddress),
		sdk.NewAttribute(types.AttrRecipient, p.RecipientAddress), sdk.NewAttribute(types.AttrAmount, p.Amount.String()),
		sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", p.Status)),
	))
	return nil
}

// --- CreateBatchPayout ---
func (k Keeper) CreateBatchPayout(ctx sdk.Context, msg *types.MsgCreateBatchPayout) error {
	params := k.GetParams(ctx)
	if !params.BatchPayoutsEnabled {
		return types.ErrBatchDisabled
	}
	if k.HasBatchPayout(ctx, msg.BatchId) {
		return fmt.Errorf("batch %s: %w", msg.BatchId, types.ErrInvalidPayoutID)
	}
	if len(msg.Payouts) == 0 || uint32(len(msg.Payouts)) > params.MaxBatchSize {
		return fmt.Errorf("batch size: %w", types.ErrInvalidPayoutID)
	}
	if err := k.validateMerchant(ctx, msg.Initiator, msg.MerchantId); err != nil {
		return err
	}

	now := ctx.BlockTime().Unix()
	seen := make(map[string]bool)
	var totalAmt, totalNet sdk.Coin
	totalFee := sdk.NewInt64Coin("unxrl", 0)
	first := true
	var payoutIDs []string

	for _, in := range msg.Payouts {
		if seen[in.PayoutId] {
			return fmt.Errorf("duplicate %s: %w", in.PayoutId, types.ErrInvalidPayoutID)
		}
		seen[in.PayoutId] = true
		if k.HasPayout(ctx, in.PayoutId) {
			return fmt.Errorf("%s exists: %w", in.PayoutId, types.ErrPayoutExists)
		}
		if in.Amount.IsZero() || in.Amount.IsNegative() {
			return fmt.Errorf("amount: %w", types.ErrAmountNotPositive)
		}
		if in.AssetDenom != "unxrl" {
			return fmt.Errorf("denom %s: v1 only supports unxrl: %w", in.AssetDenom, types.ErrInvalidDenom)
		}
		p := types.NewPayout(in.PayoutId, msg.BatchId, msg.MerchantId, msg.Initiator, in.RecipientAddress, in.AssetDenom, in.Amount, in.PayoutType, in.PayoutReference, in.Memo, now)
		if params.ApprovalRequired {
			p.Status = int32(types.PayoutCreated)
		} else {
			p.Status = int32(types.PayoutApproved)
			p.ApprovedAt = now
		}

		// Previously nothing validated individual batch entries — not the
		// recipient bech32 address, not the minimum amount, not the PayoutId
		// format, not the reference/memo length caps. A batch could silently
		// create permanently-unpayable payout records to malformed addresses.
		// This is the same per-entry validation CreatePayout already applies
		// to a single payout.
		if err := p.ValidateWithParams(params); err != nil {
			return fmt.Errorf("entry %s: %w", in.PayoutId, err)
		}

		if err := k.SetPayout(ctx, p); err != nil {
			return err
		}
		payoutIDs = append(payoutIDs, p.PayoutId)

		if first {
			totalAmt = in.Amount
			totalNet = in.Amount
			first = false
		} else {
			totalAmt = totalAmt.Add(in.Amount)
			totalNet = totalNet.Add(in.Amount)
		}
	}

	b := types.NewBatchPayout(msg.BatchId, msg.MerchantId, msg.Initiator, payoutIDs, totalAmt, totalFee, totalNet, msg.BatchReference, msg.Memo, now)
	if err := k.SetBatchPayout(ctx, b); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCreateBatch,
		sdk.NewAttribute(types.AttrBatchId, b.BatchId), sdk.NewAttribute(types.AttrInitiator, b.InitiatorAddress),
		sdk.NewAttribute(types.AttrAmount, totalAmt.String()), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", b.Status)),
	))
	return nil
}

// --- Approve ---
func (k Keeper) ApprovePayout(ctx sdk.Context, msg *types.MsgApprovePayout) error {
	p, found := k.GetPayout(ctx, msg.PayoutId)
	if !found {
		return fmt.Errorf("%s: %w", msg.PayoutId, types.ErrPayoutNotFound)
	}
	// The merchant-owner branch this used to have looked up a merchant by
	// treating p.InitiatorAddress (the requester, not a merchant key) as the
	// GetMerchant lookup address — GetMerchant is keyed by merchant owner
	// address, and p.MerchantId is a free-text label with no resolution path
	// to an address via the available keeper interface, so that branch could
	// never match and always fell through to ErrUnauthorized anyway.
	if msg.Signer != p.InitiatorAddress && msg.Signer != k.authority {
		return types.ErrUnauthorized
	}
	if p.Status != int32(types.PayoutCreated) {
		return fmt.Errorf("status %s: %w", types.PayoutStatus(p.Status), types.ErrInvalidTransition)
	}
	now := ctx.BlockTime().Unix()
	p.Status = int32(types.PayoutApproved)
	p.ApprovedAt = now
	p.UpdatedAt = now
	if err := k.SetPayout(ctx, p); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventApprovePayout,
		sdk.NewAttribute(types.AttrPayoutId, p.PayoutId), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", p.Status))))
	return nil
}

// --- MarkPaid ---
//
// Metadata-only (params.LiveEnabled=false): the payout transitions to PAID and
// records an external reference; FundsPaid stays false and no bank transfer
// occurs — unchanged from prior behaviour.
//
// Live (params.LiveEnabled=true): funds are transferred from the
// nexarail_treasury module account to the recipient BEFORE any state mutation.
// If the transfer fails, the payout state is left untouched.
func (k Keeper) MarkPayoutPaid(ctx sdk.Context, msg *types.MsgMarkPayoutPaid) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	p, found := k.GetPayout(ctx, msg.PayoutId)
	if !found {
		return fmt.Errorf("%s: %w", msg.PayoutId, types.ErrPayoutNotFound)
	}
	if p.FundsPaid {
		return fmt.Errorf("%s: %w", msg.PayoutId, types.ErrAlreadyPaid)
	}
	if p.Status != int32(types.PayoutApproved) {
		return fmt.Errorf("status %s: %w", types.PayoutStatus(p.Status), types.ErrInvalidTransition)
	}

	params := k.GetParams(ctx)
	now := ctx.BlockTime().Unix()

	// Live execution: transfer treasury -> recipient before mutating any state.
	if params.LiveEnabled {
		recipient, err := sdk.AccAddressFromBech32(p.RecipientAddress)
		if err != nil {
			return fmt.Errorf("recipient: %w", types.ErrInvalidRecipient)
		}
		payAmount := p.NetAmount
		if payAmount.IsNil() || !payAmount.IsPositive() {
			return fmt.Errorf("net amount %s: %w", payAmount, types.ErrAmountNotPositive)
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.TreasuryModuleAccount, recipient, sdk.NewCoins(payAmount)); err != nil {
			return fmt.Errorf("%w: %s", types.ErrLiveTransferFailed, err)
		}
		p.FundsPaid = true
	}

	p.Status = int32(types.PayoutPaid)
	p.PaidAt = now
	p.UpdatedAt = now
	p.ExternalReference = strings.TrimSpace(msg.ExternalReference)
	if msg.Memo != "" {
		p.Memo = strings.TrimSpace(msg.Memo)
	}
	if err := k.SetPayout(ctx, p); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventPayPayout,
		sdk.NewAttribute(types.AttrPayoutId, p.PayoutId),
		sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", p.Status)),
		sdk.NewAttribute(types.AttrFundsPaid, fmt.Sprintf("%t", p.FundsPaid)),
	))
	return nil
}

// ActivePaidPayoutTotals returns the sum of net amounts across all payouts with
// FundsPaid=true and status PAID — the total live treasury outflow via payouts.
func (k Keeper) ActivePaidPayoutTotals(ctx sdk.Context) sdk.Coins {
	totals := sdk.Coins{}
	for _, p := range k.GetAllPayouts(ctx) {
		if p.FundsPaid && p.Status == int32(types.PayoutPaid) {
			if !p.NetAmount.IsNil() && p.NetAmount.IsPositive() {
				totals = totals.Add(p.NetAmount)
			}
		}
	}
	return totals
}

// ValidatePayoutFundsInvariant enforces the live-funds safety invariant:
//   - FundsPaid=true requires status PAID;
//   - a funded payout must carry a valid recipient and a positive net amount.
func (k Keeper) ValidatePayoutFundsInvariant(ctx sdk.Context) error {
	for _, p := range k.GetAllPayouts(ctx) {
		if p.FundsPaid && p.Status != int32(types.PayoutPaid) {
			return fmt.Errorf("invariant: payout %s has FundsPaid=true but status=%s", p.PayoutId, types.PayoutStatus(p.Status))
		}
		if p.FundsPaid {
			if _, err := sdk.AccAddressFromBech32(p.RecipientAddress); err != nil {
				return fmt.Errorf("invariant: paid payout %s invalid recipient: %w", p.PayoutId, err)
			}
			if p.NetAmount.IsNil() || !p.NetAmount.IsPositive() {
				return fmt.Errorf("invariant: paid payout %s invalid net amount", p.PayoutId)
			}
		}
	}
	return nil
}

// --- Cancel ---
func (k Keeper) CancelPayout(ctx sdk.Context, msg *types.MsgCancelPayout) error {
	p, found := k.GetPayout(ctx, msg.PayoutId)
	if !found {
		return fmt.Errorf("%s: %w", msg.PayoutId, types.ErrPayoutNotFound)
	}
	if msg.Signer != p.InitiatorAddress && msg.Signer != k.authority {
		return types.ErrUnauthorized
	}
	if p.Status != int32(types.PayoutCreated) && p.Status != int32(types.PayoutApproved) {
		return fmt.Errorf("status %s: %w", types.PayoutStatus(p.Status), types.ErrInvalidTransition)
	}
	now := ctx.BlockTime().Unix()
	p.Status = int32(types.PayoutCancelled)
	p.CancelledAt = now
	p.UpdatedAt = now
	if msg.Memo != "" {
		p.Memo = strings.TrimSpace(msg.Memo)
	}
	if err := k.SetPayout(ctx, p); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCancelPayout,
		sdk.NewAttribute(types.AttrPayoutId, p.PayoutId), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", p.Status))))
	return nil
}

// --- Fail ---
func (k Keeper) FailPayout(ctx sdk.Context, msg *types.MsgFailPayout) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	p, found := k.GetPayout(ctx, msg.PayoutId)
	if !found {
		return fmt.Errorf("%s: %w", msg.PayoutId, types.ErrPayoutNotFound)
	}
	if p.Status != int32(types.PayoutCreated) && p.Status != int32(types.PayoutApproved) {
		return fmt.Errorf("status %s: %w", types.PayoutStatus(p.Status), types.ErrInvalidTransition)
	}
	now := ctx.BlockTime().Unix()
	p.Status = int32(types.PayoutFailed)
	p.FailedAt = now
	p.UpdatedAt = now
	p.FailureReason = strings.TrimSpace(msg.FailureReason)
	if err := k.SetPayout(ctx, p); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventFailPayout,
		sdk.NewAttribute(types.AttrPayoutId, p.PayoutId), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", p.Status)), sdk.NewAttribute(types.AttrReason, p.FailureReason)))
	return nil
}
