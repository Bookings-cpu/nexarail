package keeper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/nexarail/chain/x/treasury/types"
)

type Keeper struct {
	storeKey   storetypes.StoreKey
	authority  string
	bankKeeper types.BankKeeper
}

func NewKeeper(storeKey storetypes.StoreKey, authority string, bk types.BankKeeper) Keeper {
	return Keeper{storeKey: storeKey, authority: authority, bankKeeper: bk}
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
		// Matches escrow/payout/settlement: fail loud on corrupt stored params
		// rather than silently returning a half-populated Params whose nil-Int
		// coin fields panic unhelpfully somewhere downstream instead.
		panic(fmt.Errorf("treasury params: %w", err))
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
func (k Keeper) UpdateParams(ctx sdk.Context, auth string, p types.Params) error {
	if auth != k.authority {
		return types.ErrUnauthorized
	}
	return k.SetParams(ctx, p)
}

// --- Generic helpers ---
func (k Keeper) storeJSON(ctx sdk.Context, key []byte, v interface{}) {
	bz, _ := json.Marshal(v)
	ctx.KVStore(k.storeKey).Set(key, bz)
}
func (k Keeper) loadJSON(ctx sdk.Context, key []byte, dest interface{}) bool {
	bz := ctx.KVStore(k.storeKey).Get(key)
	if bz == nil {
		return false
	}
	return json.Unmarshal(bz, dest) == nil
}
func (k Keeper) has(ctx sdk.Context, key []byte) bool { return ctx.KVStore(k.storeKey).Has(key) }

// --- Treasury Accounts ---
func (k Keeper) SetTreasuryAccount(ctx sdk.Context, a types.TreasuryAccount) {
	k.storeJSON(ctx, types.AccountKey(a.AccountId), a)
}
func (k Keeper) GetTreasuryAccount(ctx sdk.Context, id string) (types.TreasuryAccount, bool) {
	var a types.TreasuryAccount
	if !k.loadJSON(ctx, types.AccountKey(id), &a) {
		return a, false
	}
	return a, true
}
func (k Keeper) HasTreasuryAccount(ctx sdk.Context, id string) bool {
	return k.has(ctx, types.AccountKey(id))
}
func (k Keeper) GetAllTreasuryAccounts(ctx sdk.Context) []types.TreasuryAccount {
	var as []types.TreasuryAccount
	k.iter(ctx, types.AccountKeyPrefix, func() interface{} { return &types.TreasuryAccount{} }, func(v interface{}) { as = append(as, *(v.(*types.TreasuryAccount))) })
	return as
}

func (k Keeper) iter(ctx sdk.Context, prefix []byte, factory func() interface{}, collect func(interface{})) {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		v := factory()
		if err := json.Unmarshal(iter.Value(), v); err != nil {
			continue
		}
		collect(v)
	}
}

// --- Budgets ---
func (k Keeper) SetBudget(ctx sdk.Context, b types.Budget) {
	k.storeJSON(ctx, types.BudgetKey(b.BudgetId), b)
	k.storeJSON(ctx, types.BudgetByAccountKey(b.AccountId, b.BudgetId), []byte{1})
}
func (k Keeper) GetBudget(ctx sdk.Context, id string) (types.Budget, bool) {
	var b types.Budget
	if !k.loadJSON(ctx, types.BudgetKey(id), &b) {
		return b, false
	}
	return b, true
}
func (k Keeper) HasBudget(ctx sdk.Context, id string) bool { return k.has(ctx, types.BudgetKey(id)) }
func (k Keeper) GetAllBudgets(ctx sdk.Context) []types.Budget {
	var bs []types.Budget
	k.iter(ctx, types.BudgetKeyPrefix, func() interface{} { return &types.Budget{} }, func(v interface{}) { bs = append(bs, *(v.(*types.Budget))) })
	return bs
}
func (k Keeper) GetBudgetsByAccount(ctx sdk.Context, acctID string) []types.Budget {
	return k.byPrefix(ctx, types.BudgetByAccountKey(acctID, ""), types.BudgetKeyPrefix, false)
}

func (k Keeper) byPrefix(ctx sdk.Context, start, primaryPrefix []byte, trimPrimary bool) []types.Budget {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), start)
	defer iter.Close()
	var bs []types.Budget
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		idStart := len(start)
		if trimPrimary {
			idStart = len(primaryPrefix)
		}
		id := string(key[idStart:])
		if b, ok := k.GetBudget(ctx, id); ok {
			bs = append(bs, b)
		}
	}
	return bs
}

// --- Grants ---
func (k Keeper) SetGrant(ctx sdk.Context, g types.Grant) {
	k.storeJSON(ctx, types.GrantKey(g.GrantId), g)
	k.storeJSON(ctx, types.GrantByBudgetKey(g.BudgetId, g.GrantId), []byte{1})
	k.storeJSON(ctx, types.GrantByRecipientKey(g.RecipientAddress, g.GrantId), []byte{1})
}
func (k Keeper) GetGrant(ctx sdk.Context, id string) (types.Grant, bool) {
	var g types.Grant
	if !k.loadJSON(ctx, types.GrantKey(id), &g) {
		return g, false
	}
	return g, true
}
func (k Keeper) HasGrant(ctx sdk.Context, id string) bool { return k.has(ctx, types.GrantKey(id)) }
func (k Keeper) GetAllGrants(ctx sdk.Context) []types.Grant {
	var gs []types.Grant
	k.iter(ctx, types.GrantKeyPrefix, func() interface{} { return &types.Grant{} }, func(v interface{}) { gs = append(gs, *(v.(*types.Grant))) })
	return gs
}
func (k Keeper) getGrantIDsByPrefix(ctx sdk.Context, start []byte) []string {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), start)
	defer iter.Close()
	var ids []string
	for ; iter.Valid(); iter.Next() {
		ids = append(ids, string(iter.Key()[len(start):]))
	}
	return ids
}
func (k Keeper) GetGrantsByBudget(ctx sdk.Context, budgetID string) []types.Grant {
	var gs []types.Grant
	for _, id := range k.getGrantIDsByPrefix(ctx, types.GrantByBudgetKey(budgetID, "")) {
		if g, ok := k.GetGrant(ctx, id); ok {
			gs = append(gs, g)
		}
	}
	return gs
}
func (k Keeper) GetGrantsByRecipient(ctx sdk.Context, recipient string) []types.Grant {
	var gs []types.Grant
	for _, id := range k.getGrantIDsByPrefix(ctx, types.GrantByRecipientKey(recipient, "")) {
		if g, ok := k.GetGrant(ctx, id); ok {
			gs = append(gs, g)
		}
	}
	return gs
}

// --- Spend Requests ---
func (k Keeper) SetSpendRequest(ctx sdk.Context, s types.SpendRequest) {
	k.storeJSON(ctx, types.SpendKey(s.SpendId), s)
	k.setSpendIdx(ctx, types.SpendByAccountKey(s.AccountId, s.SpendId))
	k.setSpendIdx(ctx, types.SpendByRequesterKey(s.RequesterAddress, s.SpendId))
	k.setSpendIdx(ctx, types.SpendByRecipientKey(s.RecipientAddress, s.SpendId))
	if s.BudgetId != "" {
		k.setSpendIdx(ctx, types.SpendByBudgetKey(s.BudgetId, s.SpendId))
	}
	if s.GrantId != "" {
		k.setSpendIdx(ctx, types.SpendByGrantKey(s.GrantId, s.SpendId))
	}
}
func (k Keeper) setSpendIdx(ctx sdk.Context, key []byte) { ctx.KVStore(k.storeKey).Set(key, []byte{1}) }
func (k Keeper) GetSpendRequest(ctx sdk.Context, id string) (types.SpendRequest, bool) {
	var s types.SpendRequest
	if !k.loadJSON(ctx, types.SpendKey(id), &s) {
		return s, false
	}
	return s, true
}
func (k Keeper) HasSpendRequest(ctx sdk.Context, id string) bool {
	return k.has(ctx, types.SpendKey(id))
}
func (k Keeper) GetAllSpendRequests(ctx sdk.Context) []types.SpendRequest {
	var ss []types.SpendRequest
	k.iter(ctx, types.SpendKeyPrefix, func() interface{} { return &types.SpendRequest{} }, func(v interface{}) { ss = append(ss, *(v.(*types.SpendRequest))) })
	return ss
}
func (k Keeper) getSpendIDsByPrefix(ctx sdk.Context, start []byte) []string {
	i := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), start)
	defer i.Close()
	var ids []string
	for ; i.Valid(); i.Next() {
		ids = append(ids, string(i.Key()[len(start):]))
	}
	return ids
}
func (k Keeper) GetSpendRequestsByAccount(ctx sdk.Context, acctID string) []types.SpendRequest {
	return k.spendsByIDs(ctx, k.getSpendIDsByPrefix(ctx, types.SpendByAccountKey(acctID, "")))
}
func (k Keeper) GetSpendRequestsByBudget(ctx sdk.Context, budgetID string) []types.SpendRequest {
	return k.spendsByIDs(ctx, k.getSpendIDsByPrefix(ctx, types.SpendByBudgetKey(budgetID, "")))
}
func (k Keeper) GetSpendRequestsByGrant(ctx sdk.Context, grantID string) []types.SpendRequest {
	return k.spendsByIDs(ctx, k.getSpendIDsByPrefix(ctx, types.SpendByGrantKey(grantID, "")))
}
func (k Keeper) GetSpendRequestsByRequester(ctx sdk.Context, req string) []types.SpendRequest {
	return k.spendsByIDs(ctx, k.getSpendIDsByPrefix(ctx, types.SpendByRequesterKey(req, "")))
}
func (k Keeper) GetSpendRequestsByRecipient(ctx sdk.Context, rec string) []types.SpendRequest {
	return k.spendsByIDs(ctx, k.getSpendIDsByPrefix(ctx, types.SpendByRecipientKey(rec, "")))
}
func (k Keeper) spendsByIDs(ctx sdk.Context, ids []string) []types.SpendRequest {
	var ss []types.SpendRequest
	for _, id := range ids {
		if s, ok := k.GetSpendRequest(ctx, id); ok {
			ss = append(ss, s)
		}
	}
	return ss
}

// --- RebuildIndexes ---
func (k Keeper) RebuildIndexes(ctx sdk.Context) {
	for _, a := range k.GetAllTreasuryAccounts(ctx) {
		k.SetTreasuryAccount(ctx, a)
	}
	for _, b := range k.GetAllBudgets(ctx) {
		k.SetBudget(ctx, b)
	}
	for _, g := range k.GetAllGrants(ctx) {
		k.SetGrant(ctx, g)
	}
	for _, s := range k.GetAllSpendRequests(ctx) {
		k.SetSpendRequest(ctx, s)
	}
}

// --- CreateTreasuryAccount ---
func (k Keeper) CreateTreasuryAccount(ctx sdk.Context, msg *types.MsgCreateTreasuryAccount) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	p := k.GetParams(ctx)
	if !p.TreasuryEnabled {
		return types.ErrTreasuryDisabled
	}
	if k.HasTreasuryAccount(ctx, msg.AccountId) {
		return fmt.Errorf("%s: %w", msg.AccountId, types.ErrRecordExists)
	}
	now := ctx.BlockTime().Unix()
	a := types.NewTreasuryAccount(msg.AccountId, msg.Category, msg.Name, msg.Description, msg.MetadataUri, msg.NominalBalance, now)
	if err := a.ValidateWithParams(p); err != nil {
		return err
	}
	k.SetTreasuryAccount(ctx, a)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCreateAccount, sdk.NewAttribute(types.AttrAccountId, a.AccountId), sdk.NewAttribute(types.AttrCategory, fmt.Sprintf("%d", a.Category)), sdk.NewAttribute(types.AttrAmount, a.NominalBalance.String())))
	return nil
}

// --- CreateBudget ---
func (k Keeper) CreateBudget(ctx sdk.Context, msg *types.MsgCreateBudget) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	p := k.GetParams(ctx)
	if !p.BudgetsEnabled {
		return types.ErrBudgetsDisabled
	}
	if !k.HasTreasuryAccount(ctx, msg.AccountId) {
		return fmt.Errorf("account %s: %w", msg.AccountId, types.ErrAccountNotFound)
	}
	if k.HasBudget(ctx, msg.BudgetId) {
		return fmt.Errorf("%s: %w", msg.BudgetId, types.ErrRecordExists)
	}
	now := ctx.BlockTime().Unix()
	b := types.NewBudget(msg.BudgetId, msg.AccountId, msg.Category, msg.Title, msg.Description, msg.TotalAmount, msg.StartTime, msg.EndTime, msg.MetadataUri, now)
	if err := b.ValidateWithParams(p); err != nil {
		return err
	}
	k.SetBudget(ctx, b)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCreateBudget, sdk.NewAttribute(types.AttrBudgetId, b.BudgetId), sdk.NewAttribute(types.AttrAmount, b.TotalAmount.String()), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", b.Status))))
	return nil
}

// --- UpdateBudgetStatus ---
func (k Keeper) UpdateBudgetStatus(ctx sdk.Context, msg *types.MsgUpdateBudgetStatus) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	b, found := k.GetBudget(ctx, msg.BudgetId)
	if !found {
		return fmt.Errorf("%s: %w", msg.BudgetId, types.ErrRecordNotFound)
	}
	if !types.IsValidBudgetStatus(msg.Status) {
		return fmt.Errorf("status %d: %w", msg.Status, types.ErrInvalidStatus)
	}
	// closed cannot reopen
	if b.Status == int32(types.BudgetClosed) && msg.Status != int32(types.BudgetClosed) {
		return fmt.Errorf("closed budgets cannot reopen: %w", types.ErrInvalidTransition)
	}
	b.Status = msg.Status
	b.UpdatedAt = ctx.BlockTime().Unix()
	k.SetBudget(ctx, b)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventUpdateBudget, sdk.NewAttribute(types.AttrBudgetId, b.BudgetId), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", b.Status))))
	return nil
}

// --- CreateGrant ---
func (k Keeper) CreateGrant(ctx sdk.Context, msg *types.MsgCreateGrant) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	p := k.GetParams(ctx)
	if !p.GrantsEnabled {
		return types.ErrGrantsDisabled
	}
	b, found := k.GetBudget(ctx, msg.BudgetId)
	if !found {
		return fmt.Errorf("budget %s: %w", msg.BudgetId, types.ErrRecordNotFound)
	}
	if b.Status != int32(types.BudgetActive) {
		return fmt.Errorf("budget %s not active: %w", msg.BudgetId, types.ErrInvalidTransition)
	}
	if k.HasGrant(ctx, msg.GrantId) {
		return fmt.Errorf("%s: %w", msg.GrantId, types.ErrRecordExists)
	}
	now := ctx.BlockTime().Unix()
	g := types.NewGrant(msg.GrantId, msg.BudgetId, msg.RecipientAddress, msg.Title, msg.Description, msg.Amount, msg.MilestoneCount, msg.MetadataUri, now)
	if err := g.ValidateWithParams(p); err != nil {
		return err
	}
	// Check budget capacity
	newAlloc := b.AllocatedAmount.Add(g.Amount)
	if newAlloc.Add(b.SpentAmount).Amount.GT(b.TotalAmount.Amount) {
		return fmt.Errorf("budget capacity exceeded: %w", types.ErrBudgetCapacity)
	}
	b.AllocatedAmount = newAlloc
	b.UpdatedAt = now
	k.SetBudget(ctx, b)
	k.SetGrant(ctx, g)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCreateGrant, sdk.NewAttribute(types.AttrGrantId, g.GrantId), sdk.NewAttribute(types.AttrAmount, g.Amount.String()), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", g.Status))))
	return nil
}

// --- UpdateGrantStatus ---
func (k Keeper) UpdateGrantStatus(ctx sdk.Context, msg *types.MsgUpdateGrantStatus) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	g, found := k.GetGrant(ctx, msg.GrantId)
	if !found {
		return fmt.Errorf("%s: %w", msg.GrantId, types.ErrRecordNotFound)
	}
	if !types.IsValidGrantStatus(msg.Status) {
		return fmt.Errorf("status %d: %w", msg.Status, types.ErrInvalidStatus)
	}
	// Terminal protection
	if g.Status == int32(types.GrantCancelled) && msg.Status != int32(types.GrantCancelled) {
		return fmt.Errorf("cancelled grant: %w", types.ErrInvalidTransition)
	}
	wasAlreadyCancelled := g.Status == int32(types.GrantCancelled)
	g.Status = msg.Status
	g.UpdatedAt = ctx.BlockTime().Unix()
	if msg.Status == int32(types.GrantCompleted) {
		g.CompletedAt = g.UpdatedAt
	}
	// Cancelling releases the budget capacity CreateGrant reserved — otherwise
	// a cancelled grant permanently consumes capacity it will never spend.
	if msg.Status == int32(types.GrantCancelled) && !wasAlreadyCancelled && g.BudgetId != "" {
		if b, found := k.GetBudget(ctx, g.BudgetId); found {
			if b.AllocatedAmount.Amount.GTE(g.Amount.Amount) {
				b.AllocatedAmount = b.AllocatedAmount.Sub(g.Amount)
			} else {
				b.AllocatedAmount = sdk.NewCoin(b.AllocatedAmount.Denom, sdk.ZeroInt())
			}
			b.UpdatedAt = g.UpdatedAt
			k.SetBudget(ctx, b)
		}
	}
	k.SetGrant(ctx, g)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventUpdateGrant, sdk.NewAttribute(types.AttrGrantId, g.GrantId), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", g.Status))))
	return nil
}

// --- CreateSpendRequest ---
func (k Keeper) CreateSpendRequest(ctx sdk.Context, msg *types.MsgCreateSpendRequest) error {
	p := k.GetParams(ctx)
	if !p.SpendRequestsEnabled {
		return types.ErrSpendDisabled
	}
	if !k.HasTreasuryAccount(ctx, msg.AccountId) {
		return fmt.Errorf("account %s: %w", msg.AccountId, types.ErrAccountNotFound)
	}
	if k.HasSpendRequest(ctx, msg.SpendId) {
		return fmt.Errorf("%s: %w", msg.SpendId, types.ErrRecordExists)
	}
	if msg.BudgetId != "" {
		if _, found := k.GetBudget(ctx, msg.BudgetId); !found {
			return fmt.Errorf("budget %s: %w", msg.BudgetId, types.ErrRecordNotFound)
		}
	}
	if msg.GrantId != "" {
		if _, found := k.GetGrant(ctx, msg.GrantId); !found {
			return fmt.Errorf("grant %s: %w", msg.GrantId, types.ErrRecordNotFound)
		}
	}
	now := ctx.BlockTime().Unix()
	s := types.NewSpendRequest(msg.SpendId, msg.AccountId, msg.BudgetId, msg.GrantId, msg.Requester, msg.RecipientAddress, msg.Amount, msg.Purpose, msg.Reference, msg.Memo, now)
	if err := s.ValidateWithParams(p); err != nil {
		return err
	}
	k.SetSpendRequest(ctx, s)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCreateSpend, sdk.NewAttribute(types.AttrSpendId, s.SpendId), sdk.NewAttribute(types.AttrAmount, s.Amount.String()), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", s.Status))))
	return nil
}

// --- ApproveSpendRequest ---
func (k Keeper) ApproveSpendRequest(ctx sdk.Context, msg *types.MsgApproveSpendRequest) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	s, found := k.GetSpendRequest(ctx, msg.SpendId)
	if !found {
		return fmt.Errorf("%s: %w", msg.SpendId, types.ErrRecordNotFound)
	}
	if s.Status != int32(types.SpendRequested) {
		return fmt.Errorf("status: %w", types.ErrInvalidTransition)
	}
	var b types.Budget
	if s.BudgetId != "" {
		var found bool
		b, found = k.GetBudget(ctx, s.BudgetId)
		if !found {
			return fmt.Errorf("budget %s: %w", s.BudgetId, types.ErrRecordNotFound)
		}
		newAlloc := b.AllocatedAmount.Add(s.Amount)
		if newAlloc.Add(b.SpentAmount).Amount.GT(b.TotalAmount.Amount) {
			return fmt.Errorf("%w", types.ErrBudgetCapacity)
		}
		// Reserve capacity now, not just check it — otherwise multiple approvals
		// against the same budget can each pass this check independently and
		// collectively overspend it (the reservation must be visible to the
		// next approval, not just computed and discarded).
		b.AllocatedAmount = newAlloc
	}
	now := ctx.BlockTime().Unix()
	s.Status = int32(types.SpendApproved)
	s.ApprovedAt = now
	s.UpdatedAt = now
	if s.BudgetId != "" {
		b.UpdatedAt = now
		k.SetBudget(ctx, b)
	}
	k.SetSpendRequest(ctx, s)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventApproveSpend, sdk.NewAttribute(types.AttrSpendId, s.SpendId), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", s.Status))))
	return nil
}

// --- RejectSpendRequest ---
func (k Keeper) RejectSpendRequest(ctx sdk.Context, msg *types.MsgRejectSpendRequest) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	s, found := k.GetSpendRequest(ctx, msg.SpendId)
	if !found {
		return fmt.Errorf("%s: %w", msg.SpendId, types.ErrRecordNotFound)
	}
	if s.Status != int32(types.SpendRequested) {
		return fmt.Errorf("status: %w", types.ErrInvalidTransition)
	}
	now := ctx.BlockTime().Unix()
	s.Status = int32(types.SpendRejected)
	s.RejectedAt = now
	s.UpdatedAt = now
	if msg.Memo != "" {
		s.Memo = strings.TrimSpace(msg.Memo)
	}
	k.SetSpendRequest(ctx, s)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventRejectSpend, sdk.NewAttribute(types.AttrSpendId, s.SpendId), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", s.Status))))
	return nil
}

// --- MarkSpendExecuted ---
func (k Keeper) MarkSpendExecuted(ctx sdk.Context, msg *types.MsgMarkSpendExecuted) error {
	if msg.Authority != k.authority {
		return types.ErrUnauthorized
	}
	s, found := k.GetSpendRequest(ctx, msg.SpendId)
	if !found {
		return fmt.Errorf("%s: %w", msg.SpendId, types.ErrRecordNotFound)
	}
	if s.Status != int32(types.SpendApproved) {
		return fmt.Errorf("status: %w", types.ErrInvalidTransition)
	}

	now := ctx.BlockTime().Unix()
	params := k.GetParams(ctx)

	// Live execution: transfer treasury module → recipient
	if params.LiveEnabled {
		recipient, err := sdk.AccAddressFromBech32(s.RecipientAddress)
		if err != nil {
			return fmt.Errorf("invalid recipient: %w", err)
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.TreasuryModuleAccount, recipient, sdk.NewCoins(s.Amount)); err != nil {
			return fmt.Errorf("live spend execution failed: %w", err)
		}
		s.FundsExecuted = true
	}

	s.Status = int32(types.SpendExecuted)
	s.ExecutedAt = now
	s.UpdatedAt = now
	if msg.Reference != "" {
		s.Reference = strings.TrimSpace(msg.Reference)
	}
	if msg.Memo != "" {
		s.Memo = strings.TrimSpace(msg.Memo)
	}
	// Move the capacity reservation from "allocated" (reserved at approval) to
	// "spent" (actually executed). Must not simply add to SpentAmount without
	// releasing the matching AllocatedAmount reservation, or the budget would
	// double-count this spend's capacity usage against future approvals.
	if s.BudgetId != "" {
		b, found := k.GetBudget(ctx, s.BudgetId)
		if !found {
			return fmt.Errorf("budget %s: %w", s.BudgetId, types.ErrRecordNotFound)
		}
		if b.AllocatedAmount.Amount.GTE(s.Amount.Amount) {
			b.AllocatedAmount = b.AllocatedAmount.Sub(s.Amount)
		} else {
			// Should not happen given Approve always reserves first, but never let
			// a subtraction panic or go negative on inconsistent legacy state.
			b.AllocatedAmount = sdk.NewCoin(b.AllocatedAmount.Denom, sdk.ZeroInt())
		}
		b.SpentAmount = b.SpentAmount.Add(s.Amount)
		b.UpdatedAt = now
		k.SetBudget(ctx, b)
	}
	k.SetSpendRequest(ctx, s)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventExecuteSpend, sdk.NewAttribute(types.AttrSpendId, s.SpendId), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", s.Status))))
	return nil
}

// ActiveExecutedSpendTotals returns the sum of amounts across all spends with FundsExecuted=true
// and status SpendExecuted. This represents the total treasury outflows via live execution.
func (k Keeper) ActiveExecutedSpendTotals(ctx sdk.Context) sdk.Coins {
	totals := sdk.Coins{}
	for _, s := range k.GetAllSpendRequests(ctx) {
		if s.FundsExecuted && s.Status == int32(types.SpendExecuted) {
			totals = totals.Add(s.Amount)
		}
	}
	return totals
}

// ValidateSpendInvariant checks budget accounting and FundsExecuted consistency
// across all spend requests. Registered as a chain invariant (see module.go).
func (k Keeper) ValidateSpendInvariant(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	for _, s := range k.GetAllSpendRequests(ctx) {
		// A spend executed while live funds were enabled must be marked as such;
		// metadata-only mode never sets FundsExecuted, so this only fires if a
		// live execution path silently failed to record the transfer it made.
		if s.Status == int32(types.SpendExecuted) && params.LiveEnabled && !s.FundsExecuted {
			return fmt.Errorf("invariant violation: spend %s executed under live_enabled but FundsExecuted=false", s.SpendId)
		}
		// Terminal spends should not have FundsExecuted without being executed
		if s.FundsExecuted && s.Status != int32(types.SpendExecuted) {
			return fmt.Errorf("invariant violation: spend %s has FundsExecuted=true but status=%s", s.SpendId, types.SpendStatus(s.Status))
		}
	}
	// Every budget's allocated+spent must stay within its ceiling — this is the
	// property the Approve/Execute/Cancel reservation fixes above are supposed
	// to maintain; catching a violation here means that accounting broke somewhere.
	for _, b := range k.GetAllBudgets(ctx) {
		if b.AllocatedAmount.Add(b.SpentAmount).Amount.GT(b.TotalAmount.Amount) {
			return fmt.Errorf("invariant violation: budget %s allocated+spent (%s) exceeds total (%s)", b.BudgetId, b.AllocatedAmount.Add(b.SpentAmount), b.TotalAmount)
		}
	}
	return nil
}

// --- CancelSpendRequest ---
func (k Keeper) CancelSpendRequest(ctx sdk.Context, msg *types.MsgCancelSpendRequest) error {
	s, found := k.GetSpendRequest(ctx, msg.SpendId)
	if !found {
		return fmt.Errorf("%s: %w", msg.SpendId, types.ErrRecordNotFound)
	}
	if msg.Signer != s.RequesterAddress && msg.Signer != k.authority {
		return types.ErrUnauthorized
	}
	if s.Status != int32(types.SpendRequested) && s.Status != int32(types.SpendApproved) {
		return fmt.Errorf("status %s: %w", types.SpendStatus(s.Status), types.ErrInvalidTransition)
	}
	// An approved-but-not-yet-executed spend holds a budget reservation from
	// ApproveSpendRequest — release it, or cancelled spends would permanently
	// consume capacity they never actually used.
	wasApproved := s.Status == int32(types.SpendApproved)
	now := ctx.BlockTime().Unix()
	s.Status = int32(types.SpendCancelled)
	s.UpdatedAt = now
	if msg.Memo != "" {
		s.Memo = strings.TrimSpace(msg.Memo)
	}
	if wasApproved && s.BudgetId != "" {
		if b, found := k.GetBudget(ctx, s.BudgetId); found {
			if b.AllocatedAmount.Amount.GTE(s.Amount.Amount) {
				b.AllocatedAmount = b.AllocatedAmount.Sub(s.Amount)
			} else {
				b.AllocatedAmount = sdk.NewCoin(b.AllocatedAmount.Denom, sdk.ZeroInt())
			}
			b.UpdatedAt = now
			k.SetBudget(ctx, b)
		}
	}
	k.SetSpendRequest(ctx, s)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventCancelSpend, sdk.NewAttribute(types.AttrSpendId, s.SpendId), sdk.NewAttribute(types.AttrStatus, fmt.Sprintf("%d", s.Status))))
	return nil
}
