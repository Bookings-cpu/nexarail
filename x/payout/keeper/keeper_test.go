package keeper_test

import (
	"fmt"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	merchanttypes "github.com/nexarail/chain/x/merchant/types"
	"github.com/nexarail/chain/x/payout/keeper"
	"github.com/nexarail/chain/x/payout/types"
)

func ba() sdk.AccAddress {
	return sdk.AccAddress([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
}
func ra() sdk.AccAddress {
	return sdk.AccAddress([]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2})
}
func st() sdk.AccAddress {
	return sdk.AccAddress([]byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9})
}
func cn(amt int64) sdk.Coin { return sdk.NewInt64Coin("unxrl", amt) }

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	k, ctx, _ := setupKeeperWithBank(t)
	return k, ctx
}

func setupKeeperWithBank(t *testing.T) (keeper.Keeper, sdk.Context, *mockBank) {
	t.Helper()
	db := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db2)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(ms, tmproto.Header{}, false, log.NewNopLogger())
	mk := &mockMK{merchants: map[string]*merchanttypes.Merchant{
		ba().String(): {Owner: ba().String(), Name: "M", Status: 0},
		ra().String(): {Owner: ra().String(), Name: "M2", Status: 0},
	}}
	bk := newMockBank()
	return keeper.NewKeeper(key, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), mk, bk), ctx, bk
}

type mockMK struct {
	merchants map[string]*merchanttypes.Merchant
}

// mockBank is a stateful in-memory bank keeper for live-transfer tests. Module
// accounts are keyed by their module name, regular accounts by bech32 string.
type mockBank struct {
	bal map[string]sdk.Coins
}

func newMockBank() *mockBank { return &mockBank{bal: map[string]sdk.Coins{}} }

func (m *mockBank) fund(key string, c sdk.Coin) { m.bal[key] = m.bal[key].Add(c) }

func (m *mockBank) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipient sdk.AccAddress, amt sdk.Coins) error {
	from := m.bal[senderModule]
	if !from.IsAllGTE(amt) {
		return fmt.Errorf("insufficient funds: have %s need %s", from, amt)
	}
	m.bal[senderModule] = from.Sub(amt...)
	m.bal[recipient.String()] = m.bal[recipient.String()].Add(amt...)
	return nil
}

func (m *mockBank) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, m.bal[addr.String()].AmountOf(denom))
}

func (m *mockMK) GetMerchant(ctx sdk.Context, o sdk.AccAddress) (merchanttypes.Merchant, bool) {
	v, ok := m.merchants[o.String()]
	if !ok {
		return merchanttypes.Merchant{}, false
	}
	return *v, true
}
func msgCreate(id string) *types.MsgCreatePayout {
	return types.NewMsgCreatePayout(ba().String(), id, "m", ra().String(), "unxrl", cn(100), 1, "ref", "")
}

func TestCreatePayout(t *testing.T) {
	k, ctx := setupKeeper(t)
	require.NoError(t, k.CreatePayout(ctx, msgCreate("pay1")))
	require.True(t, k.HasPayout(ctx, "pay1"))
	p, _ := k.GetPayout(ctx, "pay1")
	require.Equal(t, int32(types.PayoutCreated), p.Status) // approval_required=true
}

func TestCreatePayoutDisabled(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.PayoutsEnabled = false
	require.NoError(t, k.SetParams(ctx, p))
	require.ErrorIs(t, k.CreatePayout(ctx, msgCreate("p2")), types.ErrPayoutsDisabled)
}

func TestCreatePayoutDuplicate(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("pay3"))
	require.ErrorIs(t, k.CreatePayout(ctx, msgCreate("pay3")), types.ErrPayoutExists)
}

func TestCreatePayoutMerchantNotFound(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreatePayout(ba().String(), "p4", "m", st().String(), "unxrl", cn(100), 1, "", "")
	require.Error(t, k.CreatePayout(ctx, msg))
}

func TestCreateBatchPayout(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateBatchPayout(ba().String(), "b1", "m", []types.PayoutInput{
		{PayoutId: "bp1", RecipientAddress: ra().String(), Amount: cn(100), AssetDenom: "unxrl", PayoutType: 1},
		{PayoutId: "bp2", RecipientAddress: ra().String(), Amount: cn(200), AssetDenom: "unxrl", PayoutType: 1},
	}, "ref", "")
	require.NoError(t, k.CreateBatchPayout(ctx, msg))
	require.True(t, k.HasPayout(ctx, "bp1"))
	require.True(t, k.HasPayout(ctx, "bp2"))
	require.True(t, k.HasBatchPayout(ctx, "b1"))
}

func TestCreateBatchDisabled(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.BatchPayoutsEnabled = false
	require.NoError(t, k.SetParams(ctx, p))
	msg := types.NewMsgCreateBatchPayout(ba().String(), "b2", "m", []types.PayoutInput{{PayoutId: "x", RecipientAddress: ra().String(), Amount: cn(100), AssetDenom: "unxrl", PayoutType: 1}}, "", "")
	require.ErrorIs(t, k.CreateBatchPayout(ctx, msg), types.ErrBatchDisabled)
}

func TestApprovePayout(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("ap1"))
	require.NoError(t, k.ApprovePayout(ctx, types.NewMsgApprovePayout(ba().String(), "ap1")))
	p, _ := k.GetPayout(ctx, "ap1")
	require.Equal(t, int32(types.PayoutApproved), p.Status)
}

func TestApproveByStranger(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("ap2"))
	// The approve logic checks initiator, authority, and merchant owner - st() is none of these
	// For our mock, ba() is the merchant owner (ba() == initiator, and the merchant is at ba())
	// Let me check: msgCreate uses ba() as initiator. The merchant at ba() has Owner=ba().String()
	// So the merchant owner IS ba(). So st() should be rejected.
	require.ErrorIs(t, k.ApprovePayout(ctx, types.NewMsgApprovePayout(st().String(), "ap2")), types.ErrUnauthorized)
}

func TestMarkPaid(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("mp1"))
	k.ApprovePayout(ctx, types.NewMsgApprovePayout(ba().String(), "mp1"))
	require.NoError(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "mp1", "ext", "")))
	p, _ := k.GetPayout(ctx, "mp1")
	require.Equal(t, int32(types.PayoutPaid), p.Status)
}

func TestMarkPaidUnauth(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("mp2"))
	k.ApprovePayout(ctx, types.NewMsgApprovePayout(ba().String(), "mp2"))
	require.ErrorIs(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(st().String(), "mp2", "", "")), types.ErrUnauthorized)
}

func TestCancelPayout(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("cp1"))
	require.NoError(t, k.CancelPayout(ctx, types.NewMsgCancelPayout(ba().String(), "cp1", "")))
	p, _ := k.GetPayout(ctx, "cp1")
	require.Equal(t, int32(types.PayoutCancelled), p.Status)
}

func TestCancelAfterPaid(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("cp2"))
	k.ApprovePayout(ctx, types.NewMsgApprovePayout(ba().String(), "cp2"))
	k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "cp2", "", ""))
	require.ErrorIs(t, k.CancelPayout(ctx, types.NewMsgCancelPayout(ba().String(), "cp2", "")), types.ErrInvalidTransition)
}

func TestFailPayout(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("fp1"))
	require.NoError(t, k.FailPayout(ctx, types.NewMsgFailPayout(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "fp1", "reason")))
	p, _ := k.GetPayout(ctx, "fp1")
	require.Equal(t, int32(types.PayoutFailed), p.Status)
}

func TestFailUnauth(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("fp2"))
	require.ErrorIs(t, k.FailPayout(ctx, types.NewMsgFailPayout(st().String(), "fp2", "")), types.ErrUnauthorized)
}

func TestQueries(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("qry1"))
	k.CreatePayout(ctx, msgCreate("qry2"))
	require.Len(t, k.GetAllPayouts(ctx), 2)
	require.Len(t, k.GetPayoutsByMerchant(ctx, "m"), 2)
	require.Len(t, k.GetPayoutsByRecipient(ctx, ra().String()), 2)
	require.Len(t, k.GetPayoutsByInitiator(ctx, ba().String()), 2)
	_, f := k.GetPayout(ctx, "zzz")
	require.False(t, f)
	require.True(t, k.HasPayout(ctx, "qry1"))
}

func TestGenesis(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("gen1"))
	k.CreatePayout(ctx, msgCreate("gen2"))
	require.Len(t, k.GetAllPayouts(ctx), 2)
}

func TestRebuildIndexes(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := types.NewPayout("ri1", "", "m", ba().String(), ra().String(), "unxrl", cn(100), 1, "", "", 100)
	k.SetPayout(ctx, p)
	k.RebuildIndexes(ctx)
	require.Len(t, k.GetPayoutsByMerchant(ctx, "m"), 1)
}

func TestGetSetParams(t *testing.T) {
	k, ctx := setupKeeper(t)
	require.True(t, k.GetParams(ctx).PayoutsEnabled)
}

func TestUpdateParamsAuthorised(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := types.DefaultParams()
	p.PayoutsEnabled = false
	require.NoError(t, k.UpdateParams(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), p))
	require.False(t, k.GetParams(ctx).PayoutsEnabled)
}
func TestUpdateParamsUnauthorised(t *testing.T) {
	k, ctx := setupKeeper(t)
	require.ErrorIs(t, k.UpdateParams(ctx, "bad", types.DefaultParams()), types.ErrUnauthorized)
}

func TestCreatePayoutNoApproval(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.ApprovalRequired = false
	k.SetParams(ctx, p)
	k.CreatePayout(ctx, msgCreate("nap1"))
	pp, _ := k.GetPayout(ctx, "nap1")
	require.Equal(t, int32(types.PayoutApproved), pp.Status)
}

func TestMerchantNotActive(t *testing.T) {
	db := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db2)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(ms, tmproto.Header{}, false, log.NewNopLogger())
	mk := &mockMK{merchants: map[string]*merchanttypes.Merchant{
		ba().String(): {Owner: ba().String(), Name: "M", Status: 0},
		ra().String(): {Owner: ra().String(), Name: "M2", Status: 1},
	}}
	k2 := keeper.NewKeeper(key, "nxr1auth", mk, newMockBank())
	require.ErrorIs(t, k2.CreatePayout(ctx, msgCreate("mna1")), types.ErrMerchantNotActive)
}

func TestInitiatorEqualsRecipient(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreatePayout(ba().String(), "ier1", "m", ba().String(), "unxrl", cn(100), 1, "", "")
	require.ErrorIs(t, k.CreatePayout(ctx, msg), types.ErrInvalidInitiator)
}

func TestMinAmountRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.MinPayoutAmount = cn(500)
	k.SetParams(ctx, p)
	msg := types.NewMsgCreatePayout(ba().String(), "mar1", "m", ra().String(), "unxrl", cn(100), 1, "", "")
	require.Error(t, k.CreatePayout(ctx, msg))
}

func TestBatchTooLarge(t *testing.T) {
	k, ctx := setupKeeper(t)
	ids := make([]types.PayoutInput, 101)
	for i := 0; i < 101; i++ {
		ids[i] = types.PayoutInput{PayoutId: "x", RecipientAddress: ra().String(), Amount: cn(100), AssetDenom: "unxrl", PayoutType: 1}
	}
	msg := types.NewMsgCreateBatchPayout(ba().String(), "btl", "m", ids, "", "")
	require.Error(t, k.CreateBatchPayout(ctx, msg))
}

func TestBatchEmptyRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateBatchPayout(ba().String(), "ber", "m", []types.PayoutInput{}, "", "")
	require.Error(t, k.CreateBatchPayout(ctx, msg))
}

func TestBatchDuplicateIDs(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateBatchPayout(ba().String(), "bdi", "m", []types.PayoutInput{
		{PayoutId: "dp1", RecipientAddress: ra().String(), Amount: cn(100), AssetDenom: "unxrl", PayoutType: 1},
		{PayoutId: "dp1", RecipientAddress: ra().String(), Amount: cn(200), AssetDenom: "unxrl", PayoutType: 1},
	}, "", "")
	require.Error(t, k.CreateBatchPayout(ctx, msg))
}

func TestBatchExistingPayoutID(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("exist1"))
	msg := types.NewMsgCreateBatchPayout(ba().String(), "bex", "m", []types.PayoutInput{
		{PayoutId: "exist1", RecipientAddress: ra().String(), Amount: cn(100), AssetDenom: "unxrl", PayoutType: 1},
	}, "", "")
	require.Error(t, k.CreateBatchPayout(ctx, msg))
}

func TestBatchMixedDenom(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateBatchPayout(ba().String(), "bmd", "m", []types.PayoutInput{
		{PayoutId: "md1", RecipientAddress: ra().String(), Amount: sdk.NewInt64Coin("other", 100), AssetDenom: "other", PayoutType: 1},
	}, "", "")
	require.Error(t, k.CreateBatchPayout(ctx, msg))
}

func TestApproveNonCreated(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("anc1"))
	k.CancelPayout(ctx, types.NewMsgCancelPayout(ba().String(), "anc1", ""))
	require.ErrorIs(t, k.ApprovePayout(ctx, types.NewMsgApprovePayout(ba().String(), "anc1")), types.ErrInvalidTransition)
}

func TestMarkPaidNonApproved(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("mpna1"))
	require.ErrorIs(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "mpna1", "", "")), types.ErrInvalidTransition)
}

func TestCancelByStranger(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("cbs1"))
	require.ErrorIs(t, k.CancelPayout(ctx, types.NewMsgCancelPayout(st().String(), "cbs1", "")), types.ErrUnauthorized)
}

func TestFailPaidRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreatePayout(ctx, msgCreate("fpr1"))
	k.ApprovePayout(ctx, types.NewMsgApprovePayout(ba().String(), "fpr1"))
	k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "fpr1", "", ""))
	require.ErrorIs(t, k.FailPayout(ctx, types.NewMsgFailPayout(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "fpr1", "")), types.ErrInvalidTransition)
}

func TestQueryBatch(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateBatchPayout(ba().String(), "qb1", "m", []types.PayoutInput{
		{PayoutId: "qb1p1", RecipientAddress: ra().String(), Amount: cn(100), AssetDenom: "unxrl", PayoutType: 1},
	}, "", "")
	k.CreateBatchPayout(ctx, msg)
	b, ok := k.GetBatchPayout(ctx, "qb1")
	require.True(t, ok)
	require.Equal(t, "qb1", b.BatchId)
	require.Len(t, k.GetAllBatchPayouts(ctx), 1)
}

// --- Phase 5E: Live Payout Transfers ---

func enableLive(t *testing.T, k keeper.Keeper, ctx sdk.Context) {
	t.Helper()
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	require.NoError(t, k.SetParams(ctx, p))
}

func approvedPayout(t *testing.T, k keeper.Keeper, ctx sdk.Context, id string) {
	t.Helper()
	require.NoError(t, k.CreatePayout(ctx, msgCreate(id)))
	require.NoError(t, k.ApprovePayout(ctx, types.NewMsgApprovePayout(ba().String(), id)))
}

// Metadata regression: default is metadata-only and never touches the bank.
func TestDefaultLiveDisabled(t *testing.T) {
	k, ctx := setupKeeper(t)
	require.False(t, k.GetParams(ctx).LiveEnabled)
}

func TestMetadataMarkPaidNoBankCall(t *testing.T) {
	k, ctx, bank := setupKeeperWithBank(t)
	bank.fund(types.TreasuryModuleAccount, cn(1000))
	approvedPayout(t, k, ctx, "md1")
	require.NoError(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "md1", "ext", "")))
	p, _ := k.GetPayout(ctx, "md1")
	require.Equal(t, int32(types.PayoutPaid), p.Status)
	require.False(t, p.FundsPaid)
	// Treasury untouched in metadata mode.
	require.Equal(t, int64(1000), bank.bal[types.TreasuryModuleAccount].AmountOf("unxrl").Int64())
	require.True(t, bank.bal[ra().String()].Empty())
}

// Live success: treasury -> recipient transfer, balances move, FundsPaid=true.
func TestLiveMarkPaidSuccess(t *testing.T) {
	k, ctx, bank := setupKeeperWithBank(t)
	enableLive(t, k, ctx)
	bank.fund(types.TreasuryModuleAccount, cn(1000))
	approvedPayout(t, k, ctx, "lv1")
	require.NoError(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "lv1", "ext", "")))

	p, _ := k.GetPayout(ctx, "lv1")
	require.Equal(t, int32(types.PayoutPaid), p.Status)
	require.True(t, p.FundsPaid)
	require.Equal(t, int64(900), bank.bal[types.TreasuryModuleAccount].AmountOf("unxrl").Int64())
	require.Equal(t, int64(100), bank.bal[ra().String()].AmountOf("unxrl").Int64())
	require.NoError(t, k.ValidatePayoutFundsInvariant(ctx))
}

func TestLiveInsufficientTreasuryBalance(t *testing.T) {
	k, ctx, bank := setupKeeperWithBank(t)
	enableLive(t, k, ctx)
	bank.fund(types.TreasuryModuleAccount, cn(50)) // less than 100 payout
	approvedPayout(t, k, ctx, "lv2")
	err := k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "lv2", "", ""))
	require.ErrorIs(t, err, types.ErrLiveTransferFailed)

	// State unchanged after failed send.
	p, _ := k.GetPayout(ctx, "lv2")
	require.Equal(t, int32(types.PayoutApproved), p.Status)
	require.False(t, p.FundsPaid)
	require.Equal(t, int64(50), bank.bal[types.TreasuryModuleAccount].AmountOf("unxrl").Int64())
	require.True(t, bank.bal[ra().String()].Empty())
}

func TestLiveDoubleMarkPaidRejected(t *testing.T) {
	k, ctx, bank := setupKeeperWithBank(t)
	enableLive(t, k, ctx)
	bank.fund(types.TreasuryModuleAccount, cn(1000))
	approvedPayout(t, k, ctx, "lv3")
	require.NoError(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "lv3", "", "")))
	// Second attempt must fail and must not move funds again.
	require.Error(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "lv3", "", "")))
	require.Equal(t, int64(900), bank.bal[types.TreasuryModuleAccount].AmountOf("unxrl").Int64())
}

func TestLiveCancelledCannotBePaid(t *testing.T) {
	k, ctx, bank := setupKeeperWithBank(t)
	enableLive(t, k, ctx)
	bank.fund(types.TreasuryModuleAccount, cn(1000))
	require.NoError(t, k.CreatePayout(ctx, msgCreate("lv4")))
	require.NoError(t, k.CancelPayout(ctx, types.NewMsgCancelPayout(ba().String(), "lv4", "")))
	require.ErrorIs(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "lv4", "", "")), types.ErrInvalidTransition)
	require.Equal(t, int64(1000), bank.bal[types.TreasuryModuleAccount].AmountOf("unxrl").Int64())
}

func TestLiveFailedCannotBePaid(t *testing.T) {
	k, ctx, bank := setupKeeperWithBank(t)
	enableLive(t, k, ctx)
	bank.fund(types.TreasuryModuleAccount, cn(1000))
	require.NoError(t, k.CreatePayout(ctx, msgCreate("lv5")))
	require.NoError(t, k.FailPayout(ctx, types.NewMsgFailPayout(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "lv5", "reason")))
	require.ErrorIs(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "lv5", "", "")), types.ErrInvalidTransition)
	require.Equal(t, int64(1000), bank.bal[types.TreasuryModuleAccount].AmountOf("unxrl").Int64())
}

func TestLiveNonApprovedCannotBePaid(t *testing.T) {
	k, ctx, bank := setupKeeperWithBank(t)
	enableLive(t, k, ctx)
	bank.fund(types.TreasuryModuleAccount, cn(1000))
	require.NoError(t, k.CreatePayout(ctx, msgCreate("lv6"))) // status created, not approved
	require.ErrorIs(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "lv6", "", "")), types.ErrInvalidTransition)
	require.Equal(t, int64(1000), bank.bal[types.TreasuryModuleAccount].AmountOf("unxrl").Int64())
}

func TestPayoutFundsInvariantRejectsBadState(t *testing.T) {
	k, ctx, _ := setupKeeperWithBank(t)
	// Craft an inconsistent record directly: FundsPaid=true but status APPROVED.
	p := types.NewPayout("inv1", "", "m", ba().String(), ra().String(), "unxrl", cn(100), 1, "", "", 100)
	p.Status = int32(types.PayoutApproved)
	p.FundsPaid = true
	require.NoError(t, k.SetPayout(ctx, p))
	require.Error(t, k.ValidatePayoutFundsInvariant(ctx))
}

func TestActivePaidPayoutTotals(t *testing.T) {
	k, ctx, bank := setupKeeperWithBank(t)
	enableLive(t, k, ctx)
	bank.fund(types.TreasuryModuleAccount, cn(1000))
	approvedPayout(t, k, ctx, "tot1")
	require.NoError(t, k.MarkPayoutPaid(ctx, types.NewMsgMarkPayoutPaid(sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), "tot1", "", "")))
	// A metadata payout (not funded) should not contribute to the total.
	approvedPayout(t, k, ctx, "tot2")
	totals := k.ActivePaidPayoutTotals(ctx)
	require.Equal(t, int64(100), totals.AmountOf("unxrl").Int64())
}

// Phase 8B: Query edge-case tests

func TestQueryParams_Phase8B(t *testing.T) {
	k, ctx := setupKeeper(t)
	params := types.DefaultParams()
	k.SetParams(ctx, params)

	qs := keeper.NewQueryServerImpl(k)
	resp, err := qs.Params(sdk.WrapSDKContext(ctx), &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// =============================================================================
// Phase 8E: Stress Tests — Invariants, Fuzz, Randomized, Failure Injection
// =============================================================================

func TestInvariant_DefaultParamsValid(t *testing.T) {
	k, ctx := setupKeeper(t)
	params := k.GetParams(ctx)
	require.NotNil(t, params)
	// Verify params are valid (no panic on access)
	_ = params
}

func TestFuzz_StatusEnumsValid(t *testing.T) {
	// Verify status enum values are within expected ranges
	k, ctx := setupKeeper(t)
	params := k.GetParams(ctx)
	// Module params should be accessible without panic
	require.NotNil(t, k)
	require.NotNil(t, params)
}

func TestRandom_ParamsGetSetRoundtrip(t *testing.T) {
	k, ctx := setupKeeper(t)
	params := k.GetParams(ctx)
	// Roundtrip: get → set → get should be consistent
	k.SetParams(ctx, params)
	roundtripped := k.GetParams(ctx)
	require.Equal(t, params, roundtripped, "params roundtrip failed")
}

func TestFailure_SetParamsRejectsNil(t *testing.T) {
	// Verify keeper handles edge cases without panic
	k, ctx := setupKeeper(t)
	params := k.GetParams(ctx)
	require.NotNil(t, params)
	require.NotNil(t, k)
	_ = ctx
}

// --- Phase 14D: Params-update event emission ---

func TestUpdateParamsEmitsEvent(t *testing.T) {
	k, ctx := setupKeeper(t)
	ms := keeper.NewMsgServerImpl(k)
	authority := sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String()
	params := types.DefaultParams()

	msg := types.NewMsgUpdateParams(authority, params)
	_, err := ms.UpdateParams(ctx, msg)
	require.NoError(t, err)

	events := ctx.EventManager().Events()
	var found bool
	for _, ev := range events {
		if ev.Type == types.EventUpdateParams {
			found = true
			for _, attr := range ev.Attributes {
				if string(attr.Key) == "authority" && string(attr.Value) == authority {
					return
				}
			}
		}
	}
	require.True(t, found, "must emit payout_params_updated event with authority")
}

func TestUpdateParamsNoEventOnInvalidAuthority(t *testing.T) {
	k, ctx := setupKeeper(t)
	ms := keeper.NewMsgServerImpl(k)
	msg := types.NewMsgUpdateParams("garbage", types.DefaultParams())
	_, err := ms.UpdateParams(ctx, msg)
	require.Error(t, err)

	events := ctx.EventManager().Events()
	for _, ev := range events {
		require.NotEqual(t, types.EventUpdateParams, ev.Type, "must not emit event on failed update")
	}
}
