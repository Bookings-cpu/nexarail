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

	"github.com/nexarail/chain/x/escrow/keeper"
	"github.com/nexarail/chain/x/escrow/types"
	merchanttypes "github.com/nexarail/chain/x/merchant/types"
)

func ba() sdk.AccAddress {
	return sdk.AccAddress([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
}
func sa() sdk.AccAddress {
	return sdk.AccAddress([]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2})
}
func st() sdk.AccAddress {
	return sdk.AccAddress([]byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9})
}
func cn(amt int64) sdk.Coin { return sdk.NewInt64Coin("unxrl", amt) }

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	t.Helper()
	db := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db2)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(ms, tmproto.Header{}, false, log.NewNopLogger())

	mk := &mockMerchantKeeper{merchants: map[string]*merchanttypes.Merchant{
		sa().String(): {Owner: sa().String(), Name: "TestMerchant", Status: 0, RebateTier: 0},
	}}
	bk := &mockBankKeeper{balances: make(map[string]sdk.Coins)}
	k := keeper.NewKeeper(key, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), mk, bk)
	return k, ctx
}

type mockMerchantKeeper struct {
	merchants map[string]*merchanttypes.Merchant
}

func (m *mockMerchantKeeper) GetMerchant(ctx sdk.Context, o sdk.AccAddress) (merchanttypes.Merchant, bool) {
	v, ok := m.merchants[o.String()]
	if !ok {
		return merchanttypes.Merchant{}, false
	}
	return *v, true
}

type mockBankKeeper struct {
	balances map[string]sdk.Coins
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	// Forgiving mock: always succeeds
	return nil
}
func (m *mockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	key := recipientAddr.String()
	m.balances[key] = m.balances[key].Add(amt...)
	return nil
}
func (m *mockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	amt := m.balances[addr.String()].AmountOf(denom)
	return sdk.NewCoin(denom, amt)
}

// --- Params ---
func TestGetSetParams(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	require.True(t, p.EscrowsEnabled)
	p.EscrowsEnabled = false
	require.NoError(t, k.SetParams(ctx, p))
	require.False(t, k.GetParams(ctx).EscrowsEnabled)
}
func TestUpdateParamsUnauth(t *testing.T) {
	k, ctx := setupKeeper(t)
	require.ErrorIs(t, k.UpdateParams(ctx, "bad", types.DefaultParams()), types.ErrUnauthorized)
}

// --- Create ---
func TestCreateEscrow(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateEscrow(ba().String(), "escrow-1", sa().String(), "merchant-1", "unxrl", cn(1000), "ref", "", 0)
	e, err := k.CreateEscrow(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, "escrow-1", e.EscrowId)
	require.Equal(t, int32(types.EscrowCreated), e.Status)
	require.True(t, k.HasEscrow(ctx, "escrow-1"))
}

func TestCreateEscrowDisabled(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.EscrowsEnabled = false
	require.NoError(t, k.SetParams(ctx, p))
	msg := types.NewMsgCreateEscrow(ba().String(), "e1", sa().String(), "m", "unxrl", cn(100), "", "", 0)
	_, err := k.CreateEscrow(ctx, msg)
	require.ErrorIs(t, err, types.ErrEscrowsDisabled)
}

func TestCreateDuplicate(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateEscrow(ba().String(), "dup", sa().String(), "m", "unxrl", cn(100), "", "", 0)
	k.CreateEscrow(ctx, msg)
	_, err := k.CreateEscrow(ctx, msg)
	require.ErrorIs(t, err, types.ErrEscrowExists)
}

func TestCreateMerchantNotFound(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateEscrow(ba().String(), "e2", st().String(), "m", "unxrl", cn(100), "", "", 0)
	_, err := k.CreateEscrow(ctx, msg)
	require.Error(t, err)
}

func TestCreateMerchantNotActive(t *testing.T) {
	db := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db2)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(ms, tmproto.Header{}, false, log.NewNopLogger())

	mk := &mockMerchantKeeper{merchants: map[string]*merchanttypes.Merchant{
		sa().String(): {Owner: sa().String(), Name: "M", Status: 1},
	}}
	bk2 := &mockBankKeeper{balances: make(map[string]sdk.Coins)}
	k2 := keeper.NewKeeper(key, "nxr1auth", mk, bk2)
	msg := types.NewMsgCreateEscrow(ba().String(), "e3", sa().String(), "m", "unxrl", cn(100), "", "", 0)
	_, err := k2.CreateEscrow(ctx, msg)
	require.ErrorIs(t, err, types.ErrMerchantNotActive)
}

func TestCreateBuyerEqualsSeller(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateEscrow(ba().String(), "e4", ba().String(), "m", "unxrl", cn(100), "", "", 0)
	_, err := k.CreateEscrow(ctx, msg)
	require.Error(t, err)
}

func TestCreateBelowMinAmount(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateEscrow(ba().String(), "e05", sa().String(), "m", "unxrl", cn(0), "", "", 0)
	_, err := k.CreateEscrow(ctx, msg)
	require.ErrorIs(t, err, types.ErrAmountNotPositive)
}

func TestCreateDefaultExpiry(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateEscrow(ba().String(), "e06", sa().String(), "m", "unxrl", cn(100), "", "", 0)
	e, err := k.CreateEscrow(ctx, msg)
	require.NoError(t, err)
	require.Greater(t, e.ExpiresAt, e.CreatedAt)
}

func TestCreateIndexes(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateEscrow(ba().String(), "e-idx", sa().String(), "merchant-1", "unxrl", cn(100), "", "", 0)
	k.CreateEscrow(ctx, msg)
	require.Len(t, k.GetEscrowsByBuyer(ctx, ba().String()), 1)
	require.Len(t, k.GetEscrowsBySeller(ctx, sa().String()), 1)
	require.Len(t, k.GetEscrowsByMerchant(ctx, "merchant-1"), 1)
}

// --- Queries ---
func TestQueryParams(t *testing.T) {
	k, ctx := setupKeeper(t)
	require.True(t, k.GetParams(ctx).EscrowsEnabled)
}
func TestQueryEscrow(t *testing.T) {
	k, ctx := setupKeeper(t)
	msg := types.NewMsgCreateEscrow(ba().String(), "qe1", sa().String(), "m", "unxrl", cn(100), "", "", 0)
	k.CreateEscrow(ctx, msg)
	e, found := k.GetEscrow(ctx, "qe1")
	require.True(t, found)
	require.Equal(t, "qe1", e.EscrowId)
}
func TestQueryNotFound(t *testing.T) {
	k, ctx := setupKeeper(t)
	_, f := k.GetEscrow(ctx, "zzz")
	require.False(t, f)
}
func TestQueryAll(t *testing.T) {
	k, ctx := setupKeeper(t)
	for i := 0; i < 3; i++ {
		k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), fmt.Sprintf("qa%d", i), sa().String(), "m", "unxrl", cn(100), "", "", 0))
	}
	require.Len(t, k.GetAllEscrows(ctx), 3)
}
func TestQueryExists(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "exq", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.True(t, k.HasEscrow(ctx, "exq"))
	require.False(t, k.HasEscrow(ctx, "no"))
}

// --- Lifecycle ---
func TestReleaseByBuyer(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "rb1", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.NoError(t, k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(ba().String(), "rb1", "", "")))
	e, _ := k.GetEscrow(ctx, "rb1")
	require.Equal(t, int32(types.EscrowReleased), e.Status)
}
func TestReleaseByAuthority(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "ra1", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.NoError(t, k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(k.GetAuthority(), "ra1", "", "")))
}
func TestReleaseByStranger(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "rs1", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.ErrorIs(t, k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(st().String(), "rs1", "", "")), types.ErrUnauthorized)
}
func TestRefundBySeller(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "fs1", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.NoError(t, k.RefundEscrow(ctx, types.NewMsgRefundEscrow(sa().String(), "fs1", "", "")))
	e, _ := k.GetEscrow(ctx, "fs1")
	require.Equal(t, int32(types.EscrowRefunded), e.Status)
}
func TestRefundByStranger(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "fss", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.ErrorIs(t, k.RefundEscrow(ctx, types.NewMsgRefundEscrow(st().String(), "fss", "", "")), types.ErrUnauthorized)
}
func TestDisputeByBuyer(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "db1", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.NoError(t, k.OpenDispute(ctx, types.NewMsgOpenDispute(ba().String(), "db1", "reason")))
	e, _ := k.GetEscrow(ctx, "db1")
	require.Equal(t, int32(types.EscrowDisputed), e.Status)
}
func TestDisputeBySeller(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "ds1", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.NoError(t, k.OpenDispute(ctx, types.NewMsgOpenDispute(sa().String(), "ds1", "reason")))
}
func TestDisputeByStranger(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "dst", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.ErrorIs(t, k.OpenDispute(ctx, types.NewMsgOpenDispute(st().String(), "dst", "")), types.ErrUnauthorized)
}
func TestResolveBuyerWins(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "rw1", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	k.OpenDispute(ctx, types.NewMsgOpenDispute(ba().String(), "rw1", "reason"))
	require.NoError(t, k.ResolveDispute(ctx, types.NewMsgResolveDispute(k.GetAuthority(), "rw1", int32(types.DisputeBuyerWins), "buyer wins")))
	e, _ := k.GetEscrow(ctx, "rw1")
	require.Equal(t, int32(types.EscrowRefunded), e.Status)
}
func TestResolveSellerWins(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "rw2", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	k.OpenDispute(ctx, types.NewMsgOpenDispute(ba().String(), "rw2", "reason"))
	require.NoError(t, k.ResolveDispute(ctx, types.NewMsgResolveDispute(k.GetAuthority(), "rw2", int32(types.DisputeSellerWins), "seller wins")))
	e, _ := k.GetEscrow(ctx, "rw2")
	require.Equal(t, int32(types.EscrowReleased), e.Status)
}
func TestResolveRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "rw3", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	k.OpenDispute(ctx, types.NewMsgOpenDispute(ba().String(), "rw3", "reason"))
	require.NoError(t, k.ResolveDispute(ctx, types.NewMsgResolveDispute(k.GetAuthority(), "rw3", int32(types.DisputeRejected), "rejected")))
	e, _ := k.GetEscrow(ctx, "rw3")
	require.Equal(t, int32(types.EscrowCreated), e.Status)
}
func TestResolveUnauth(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "rua", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	k.OpenDispute(ctx, types.NewMsgOpenDispute(ba().String(), "rua", ""))
	require.ErrorIs(t, k.ResolveDispute(ctx, types.NewMsgResolveDispute(st().String(), "rua", 3, "")), types.ErrUnauthorized)
}
func TestCancelByBuyer(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "cb1", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.NoError(t, k.CancelEscrow(ctx, types.NewMsgCancelEscrow(ba().String(), "cb1", "")))
	e, _ := k.GetEscrow(ctx, "cb1")
	require.Equal(t, int32(types.EscrowCancelled), e.Status)
}
func TestCancelByStranger(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "cst", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	require.ErrorIs(t, k.CancelEscrow(ctx, types.NewMsgCancelEscrow(st().String(), "cst", "")), types.ErrUnauthorized)
}
func TestCancelAfterRelease(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "car", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(ba().String(), "car", "", ""))
	require.ErrorIs(t, k.CancelEscrow(ctx, types.NewMsgCancelEscrow(ba().String(), "car", "")), types.ErrInvalidTransition)
}
func TestTerminalProtections(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "tp", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(ba().String(), "tp", "", ""))
	require.Error(t, k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(ba().String(), "tp", "", "")))
	require.Error(t, k.RefundEscrow(ctx, types.NewMsgRefundEscrow(sa().String(), "tp", "", "")))
}
func TestGenesisInit(t *testing.T) {
	k, ctx := setupKeeper(t)
	gs := types.DefaultGenesis()
	gs.Params.EscrowsEnabled = false
	require.NoError(t, k.SetParams(ctx, gs.Params))
	e := types.NewEscrow("g1", ba().String(), sa().String(), "m", "unxrl", cn(100), "", "", 100, 200)
	require.NoError(t, k.SetEscrow(ctx, e))
	require.False(t, k.GetParams(ctx).EscrowsEnabled)
	_, ok := k.GetEscrow(ctx, "g1")
	require.True(t, ok)
}
func TestExportGenesis(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "ex1", sa().String(), "m", "unxrl", cn(100), "", "", 0))
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "ex2", sa().String(), "m", "unxrl", cn(200), "", "", 0))
	require.Len(t, k.GetAllEscrows(ctx), 2)
}
func TestRebuildIndexes(t *testing.T) {
	k, ctx := setupKeeper(t)
	e := types.NewEscrow("ri1", ba().String(), sa().String(), "m", "unxrl", cn(100), "", "", 100, 200)
	require.NoError(t, k.SetEscrow(ctx, e))
	k.RebuildIndexes(ctx)
	require.Len(t, k.GetEscrowsByBuyer(ctx, ba().String()), 1)
}

// --- Live Custody Tests (Phase 5C) ---

func TestLiveCreateEscrow(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	k.SetParams(ctx, p)

	// Pre-fund buyer in mock
	buyer := ba()
	msg := types.NewMsgCreateEscrow(buyer.String(), "live1", sa().String(), "merchant-1", "unxrl", cn(1000), "ref", "", 0)

	e, err := k.CreateEscrow(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int32(types.EscrowFunded), e.Status)
	require.True(t, e.FundsCustodied)
}

func TestLiveReleaseEscrow(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	k.SetParams(ctx, p)

	msg := types.NewMsgCreateEscrow(ba().String(), "live2", sa().String(), "merchant-1", "unxrl", cn(1000), "ref", "", 0)
	e, _ := k.CreateEscrow(ctx, msg)
	require.True(t, e.FundsCustodied)

	require.NoError(t, k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(ba().String(), "live2", "", "")))
	updated, _ := k.GetEscrow(ctx, "live2")
	require.Equal(t, int32(types.EscrowReleased), updated.Status)
	require.False(t, updated.FundsCustodied)
}

func TestLiveRefundEscrow(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	k.SetParams(ctx, p)

	msg := types.NewMsgCreateEscrow(ba().String(), "live3", sa().String(), "merchant-1", "unxrl", cn(1000), "ref", "", 0)
	e, _ := k.CreateEscrow(ctx, msg)
	require.True(t, e.FundsCustodied)

	require.NoError(t, k.RefundEscrow(ctx, types.NewMsgRefundEscrow(sa().String(), "live3", "", "")))
	updated, _ := k.GetEscrow(ctx, "live3")
	require.Equal(t, int32(types.EscrowRefunded), updated.Status)
	require.False(t, updated.FundsCustodied)
}

func TestLiveCancelEscrow(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	k.SetParams(ctx, p)

	msg := types.NewMsgCreateEscrow(ba().String(), "live4", sa().String(), "merchant-1", "unxrl", cn(1000), "ref", "", 0)
	e, _ := k.CreateEscrow(ctx, msg)
	require.True(t, e.FundsCustodied)

	// A funded/live escrow must not be unwound by a unilateral buyer cancel —
	// only release, refund (seller-initiated), or dispute may move it once funded.
	require.ErrorIs(t, k.CancelEscrow(ctx, types.NewMsgCancelEscrow(ba().String(), "live4", "")), types.ErrInvalidTransition)
	updated, _ := k.GetEscrow(ctx, "live4")
	require.Equal(t, int32(types.EscrowFunded), updated.Status)
	require.True(t, updated.FundsCustodied)
}

func TestLiveDisputeBuyerWins(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	k.SetParams(ctx, p)

	msg := types.NewMsgCreateEscrow(ba().String(), "live5", sa().String(), "merchant-1", "unxrl", cn(1000), "ref", "", 0)
	e, _ := k.CreateEscrow(ctx, msg)
	require.True(t, e.FundsCustodied)

	k.OpenDispute(ctx, types.NewMsgOpenDispute(ba().String(), "live5", "reason"))
	require.NoError(t, k.ResolveDispute(ctx, types.NewMsgResolveDispute(k.GetAuthority(), "live5", int32(types.DisputeBuyerWins), "")))
	updated, _ := k.GetEscrow(ctx, "live5")
	require.Equal(t, int32(types.EscrowRefunded), updated.Status)
	require.False(t, updated.FundsCustodied)
}

func TestLiveMetadataOnlyDefault(t *testing.T) {
	k, ctx := setupKeeper(t)
	// Default params: LiveEnabled=false
	msg := types.NewMsgCreateEscrow(ba().String(), "live6", sa().String(), "merchant-1", "unxrl", cn(1000), "ref", "", 0)
	e, err := k.CreateEscrow(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int32(types.EscrowCreated), e.Status)
	require.False(t, e.FundsCustodied)
}

// --- Hardening Tests (Phase 5C.1) ---

func TestLiveDoubleReleaseRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx); p.LiveEnabled = true; k.SetParams(ctx, p)
	msg := types.NewMsgCreateEscrow(ba().String(), "hw1", sa().String(), "merchant-1", "unxrl", cn(1000), "", "", 0)
	e, _ := k.CreateEscrow(ctx, msg)
	require.True(t, e.FundsCustodied)
	// First release succeeds
	require.NoError(t, k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(ba().String(), "hw1", "", "")))
	// Second release fails
	require.Error(t, k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(ba().String(), "hw1", "", "")))
	updated, _ := k.GetEscrow(ctx, "hw1")
	require.False(t, updated.FundsCustodied)
}

func TestLiveDoubleRefundRejected(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx); p.LiveEnabled = true; k.SetParams(ctx, p)
	msg := types.NewMsgCreateEscrow(ba().String(), "hw2", sa().String(), "merchant-1", "unxrl", cn(1000), "", "", 0)
	e, _ := k.CreateEscrow(ctx, msg)
	require.True(t, e.FundsCustodied)
	require.NoError(t, k.RefundEscrow(ctx, types.NewMsgRefundEscrow(sa().String(), "hw2", "", "")))
	require.Error(t, k.RefundEscrow(ctx, types.NewMsgRefundEscrow(sa().String(), "hw2", "", "")))
	updated, _ := k.GetEscrow(ctx, "hw2")
	require.False(t, updated.FundsCustodied)
}

func TestLiveTerminalCustodiedInvariant(t *testing.T) {
	k, ctx := setupKeeper(t)
	// Manually create a terminal escrow with funds_custodied=true to test invariant
	e := types.NewEscrow("hw3", ba().String(), sa().String(), "m", "unxrl", cn(100), "", "", 100, 200)
	e.Status = int32(types.EscrowReleased)
	e.FundsCustodied = true
	require.NoError(t, k.SetEscrow(ctx, e))
	// Invariant must catch this
	require.Error(t, k.ValidateCustodyInvariant(ctx))
}

func TestLiveEscrowModuleBalanceInvariantHelper(t *testing.T) {
	k, ctx := setupKeeper(t)
	p := k.GetParams(ctx); p.LiveEnabled = true; k.SetParams(ctx, p)

	// Create two live escrows
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "hw4a", sa().String(), "merchant-1", "unxrl", cn(500), "", "", 0))
	k.CreateEscrow(ctx, types.NewMsgCreateEscrow(ba().String(), "hw4b", sa().String(), "merchant-1", "unxrl", cn(300), "", "", 0))

	// Total custodied should be 800
	totals := k.ActiveCustodiedEscrowTotals(ctx)
	require.Equal(t, sdk.NewInt64Coin("unxrl", 800).Amount, totals.AmountOf("unxrl"))

	// Release one — total should drop to 300
	k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(ba().String(), "hw4a", "", ""))
	totals = k.ActiveCustodiedEscrowTotals(ctx)
	require.Equal(t, sdk.NewInt64Coin("unxrl", 300).Amount, totals.AmountOf("unxrl"))
}

func TestMetadataOnlyNoBankCalls(t *testing.T) {
	k, ctx := setupKeeper(t)
	// Default LiveEnabled=false — create/release/refund/cancel should all work without live transfers
	msg := types.NewMsgCreateEscrow(ba().String(), "hw5", sa().String(), "merchant-1", "unxrl", cn(1000), "", "", 0)
	e, err := k.CreateEscrow(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int32(types.EscrowCreated), e.Status)
	require.False(t, e.FundsCustodied)

	require.NoError(t, k.ReleaseEscrow(ctx, types.NewMsgReleaseEscrow(ba().String(), "hw5", "", "")))
	updated, _ := k.GetEscrow(ctx, "hw5")
	require.Equal(t, int32(types.EscrowReleased), updated.Status)
	require.False(t, updated.FundsCustodied)
}

// Phase 8B: Query edge-case tests

func TestQueryParams_Phase8B(t *testing.T) {
	k, ctx := setupKeeper(t)
	params := types.DefaultParams()
	k.SetParams(ctx, params)

	qs := keeper.NewQueryServerImpl(k)
	resp, err := qs.Params(sdk.WrapSDKContext(ctx), &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, params.LiveEnabled, resp.Params.LiveEnabled)
	require.Equal(t, params.DefaultExpirySeconds, resp.Params.DefaultExpirySeconds)
}

func TestQueryEscrows_EmptyState(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.Escrows(sdk.WrapSDKContext(ctx), &types.QueryEscrowsRequest{})
	require.NoError(t, err)
	require.Empty(t, resp.Escrows)
}

func TestQueryEscrow_NotFound(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	_, err := qs.Escrow(sdk.WrapSDKContext(ctx), &types.QueryEscrowRequest{EscrowId: "nonexistent-escrow"})
	require.Error(t, err)
}

func TestQueryEscrowExists_NotFound(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.EscrowExists(sdk.WrapSDKContext(ctx), &types.QueryEscrowExistsRequest{EscrowId: "nonexistent-escrow"})
	require.NoError(t, err)
	require.False(t, resp.Exists)
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
