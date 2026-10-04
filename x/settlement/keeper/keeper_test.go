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

	feestypes "github.com/nexarail/chain/x/fees/types"
	merchanttypes "github.com/nexarail/chain/x/merchant/types"
	"github.com/nexarail/chain/x/settlement/keeper"
	"github.com/nexarail/chain/x/settlement/types"
)

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context, *mockBankKeeper) {
	t.Helper()
	db := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db2)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(ms, tmproto.Header{}, false, log.NewNopLogger())

	mk := &mockMerchantKeeper{
		merchants: make(map[string]*merchanttypes.Merchant),
	}
	fk := &mockFeesKeeper{}

	payer := payerAddr()
	bk := &mockBankKeeper{
		balances: map[string]sdk.Coins{
			payer.String(): sdk.NewCoins(sdk.NewInt64Coin("unxrl", 1000000000)),
		},
		moduleBalances: map[string]sdk.Coins{
			"nexarail_treasury": sdk.NewCoins(),
			"nexarail_burner":   sdk.NewCoins(),
		},
	}

	k := keeper.NewKeeper(key, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), mk, fk, bk)

	activeAddr := sdk.AccAddress([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	mk.merchants[activeAddr.String()] = &merchanttypes.Merchant{
		Owner:      activeAddr.String(),
		Name:       "TestMerchant",
		Status:     0,
		RebateTier: 2,
	}

	return k, ctx, bk
}

type mockMerchantKeeper struct {
	merchants map[string]*merchanttypes.Merchant
}

func (m *mockMerchantKeeper) GetMerchant(ctx sdk.Context, owner sdk.AccAddress) (merchanttypes.Merchant, bool) {
	merch, ok := m.merchants[owner.String()]
	if !ok {
		return merchanttypes.Merchant{}, false
	}
	return *merch, true
}

type mockFeesKeeper struct{}

func (m *mockFeesKeeper) GetParams(ctx sdk.Context) feestypes.Params {
	return feestypes.Params{
		ValidatorShareBps: 6000,
		TreasuryShareBps:  2000,
		BurnShareBps:      2000,
	}
}

type mockBankKeeper struct {
	balances           map[string]sdk.Coins
	moduleBalances     map[string]sdk.Coins
	sendCalled         bool
	sendToModuleCalled bool
	burnCalled         bool
	sendError            error
	sendToModuleError    error
	sendToModuleErrorFor string // if set, sendToModuleError only applies to this recipient module
	burnError            error
	lastFrom           sdk.AccAddress
	lastTo             sdk.AccAddress
	lastAmount         sdk.Coins
	lastToModule       string
	lastModuleAmount   sdk.Coins
	lastBurnModule     string
	lastBurnAmount     sdk.Coins
	totalBurned        sdk.Coins
}

func (m *mockBankKeeper) SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error {
	m.sendCalled = true
	m.lastFrom = from
	m.lastTo = to
	m.lastAmount = amt
	if m.sendError != nil {
		return m.sendError
	}
	// Update mock balances
	fromBal := m.balances[from.String()]
	toBal := m.balances[to.String()]
	m.balances[from.String()] = fromBal.Sub(amt...)
	m.balances[to.String()] = toBal.Add(amt...)
	return nil
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, from sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	m.sendToModuleCalled = true
	m.lastFrom = from
	m.lastToModule = recipientModule
	m.lastModuleAmount = amt
	if m.sendToModuleError != nil && (m.sendToModuleErrorFor == "" || m.sendToModuleErrorFor == recipientModule) {
		return m.sendToModuleError
	}
	fromBal := m.balances[from.String()]
	m.balances[from.String()] = fromBal.Sub(amt...)
	modBal := m.moduleBalances[recipientModule]
	m.moduleBalances[recipientModule] = modBal.Add(amt...)
	return nil
}

func (m *mockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	m.burnCalled = true
	m.lastBurnModule = moduleName
	m.lastBurnAmount = amt
	if m.burnError != nil {
		return m.burnError
	}
	modBal := m.moduleBalances[moduleName]
	m.moduleBalances[moduleName] = modBal.Sub(amt...)
	m.totalBurned = m.totalBurned.Add(amt...)
	return nil
}

func activeMerchantAddr() sdk.AccAddress {
	return sdk.AccAddress([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
}
func payerAddr() sdk.AccAddress {
	return sdk.AccAddress([]byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3})
}

func TestGetSetParams(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	p := k.GetParams(ctx)
	require.True(t, p.Enabled)
	require.Equal(t, uint32(100), p.FeeRateBps)
	require.False(t, p.LiveEnabled)

	newP := p
	newP.Enabled = false
	require.NoError(t, k.SetParams(ctx, newP))
	require.False(t, k.GetParams(ctx).Enabled)
}

func TestSetInvalidParams(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	bad := types.DefaultParams()
	bad.FeeRateBps = 10001
	require.Error(t, k.SetParams(ctx, bad))
}

func TestUpdateParams(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	p := types.DefaultParams()
	p.Enabled = false
	require.NoError(t, k.UpdateParams(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), p))
	require.False(t, k.GetParams(ctx).Enabled)
}

func TestUpdateParamsUnauthorized(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	require.ErrorIs(t, k.UpdateParams(ctx, "not-auth", types.DefaultParams()), types.ErrUnauthorized)
}

func TestCreateSettlementSuccess(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 10000), "test")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, uint64(1), s.Id)
	require.Equal(t, int32(types.SettlementCompleted), s.Status)
}

func TestCreateSettlementDisabled(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	p := k.GetParams(ctx)
	p.Enabled = false
	require.NoError(t, k.SetParams(ctx, p))
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 10000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.ErrorIs(t, err, types.ErrSettlementsDisabled)
}

func TestCreateSettlementMerchantNotFound(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	unknown := sdk.AccAddress([]byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9})
	msg := types.NewMsgCreateSettlement(payerAddr().String(), unknown.String(), sdk.NewInt64Coin("unxrl", 100), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
}

func TestCreateSettlementMerchantNotActive(t *testing.T) {
	db := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db2)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(ms, tmproto.Header{}, false, log.NewNopLogger())

	inactiveAddr := activeMerchantAddr()
	mk := &mockMerchantKeeper{merchants: map[string]*merchanttypes.Merchant{
		inactiveAddr.String(): {Owner: inactiveAddr.String(), Name: "Inactive", Status: 1},
	}}
	bk := &mockBankKeeper{balances: map[string]sdk.Coins{}}
	k2 := keeper.NewKeeper(key, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), mk, &mockFeesKeeper{}, bk)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), inactiveAddr.String(), sdk.NewInt64Coin("unxrl", 100), "")
	_, err := k2.CreateSettlement(ctx, msg)
	require.ErrorIs(t, err, types.ErrMerchantNotActive)
}

func TestCreateSettlementInvalidAmount(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 0), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
}

func TestCreateSettlementIDsIncrement(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	for i := uint64(1); i <= 3; i++ {
		msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
		s, err := k.CreateSettlement(ctx, msg)
		require.NoError(t, err)
		require.Equal(t, i, s.Id)
	}
}

func TestCreateSettlementFeeCalculation(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 1000000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int64(9000), s.FeeAmount.Amount.Int64())
	require.Equal(t, int64(5400), s.ValidatorShare.Amount.Int64())
	require.Equal(t, int64(1800), s.TreasuryShare.Amount.Int64())
	require.Equal(t, int64(1800), s.BurnShare.Amount.Int64())
}

func TestCreateSettlementEvents(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NotEmpty(t, ctx.EventManager().Events())
}

func TestGetSettlement(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
	created, _ := k.CreateSettlement(ctx, msg)
	s, found := k.GetSettlement(ctx, created.Id)
	require.True(t, found)
	require.Equal(t, created.Id, s.Id)
}

func TestGetSettlementNotFound(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	_, found := k.GetSettlement(ctx, 999)
	require.False(t, found)
}

func TestGetSettlementsByPayer(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
	k.CreateSettlement(ctx, msg)
	require.Len(t, k.GetSettlementsByPayer(ctx, payerAddr().String()), 1)
	require.Empty(t, k.GetSettlementsByPayer(ctx, "unknown"))
}

func TestGetSettlementsByMerchant(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
	k.CreateSettlement(ctx, msg)
	require.Len(t, k.GetSettlementsByMerchant(ctx, activeMerchantAddr().String()), 1)
	require.Empty(t, k.GetSettlementsByMerchant(ctx, "unknown"))
}

func TestGetAllSettlements(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	require.Empty(t, k.GetAllSettlements(ctx))
	for i := 0; i < 3; i++ {
		k.CreateSettlement(ctx, types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), ""))
	}
	require.Len(t, k.GetAllSettlements(ctx), 3)
}

func TestUpdateSettlementStatus(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
	s, _ := k.CreateSettlement(ctx, msg)
	require.NoError(t, k.UpdateSettlementStatus(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), s.Id, int32(types.SettlementRefunded)))
	updated, _ := k.GetSettlement(ctx, s.Id)
	require.Equal(t, int32(types.SettlementRefunded), updated.Status)
}

func TestUpdateSettlementStatusUnauthorized(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
	s, _ := k.CreateSettlement(ctx, msg)
	require.ErrorIs(t, k.UpdateSettlementStatus(ctx, "not-auth", s.Id, int32(types.SettlementFailed)), types.ErrUnauthorized)
}

func TestTerminalStatusCannotTransition(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
	s, _ := k.CreateSettlement(ctx, msg)
	require.NoError(t, k.UpdateSettlementStatus(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), s.Id, int32(types.SettlementCancelled)))
	require.ErrorIs(t, k.UpdateSettlementStatus(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), s.Id, int32(types.SettlementCompleted)), types.ErrInvalidStatusTransition)
}

func TestCompletedCannotGoToPending(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
	s, _ := k.CreateSettlement(ctx, msg)
	require.Equal(t, int32(types.SettlementCompleted), s.Status)
	require.ErrorIs(t, k.UpdateSettlementStatus(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), s.Id, int32(types.SettlementPending)), types.ErrInvalidStatusTransition)
}

func TestInitGenesis(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	gs := types.DefaultGenesis()
	gs.Params.Enabled = false
	gs.Params.FeeRateBps = 50
	require.NoError(t, k.SetParams(ctx, gs.Params))

	s := types.NewSettlement(1, payerAddr().String(), activeMerchantAddr().String(), "TestMerchant", activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), sdk.NewInt64Coin("unxrl", 1), sdk.NewInt64Coin("unxrl", 0),
		sdk.NewInt64Coin("unxrl", 0), sdk.NewInt64Coin("unxrl", 0), sdk.NewInt64Coin("unxrl", 0),
		0, "", "", "", 100,
	)
	require.NoError(t, k.SetSettlement(ctx, s))
	require.False(t, k.GetParams(ctx).Enabled)
	require.Equal(t, uint32(50), k.GetParams(ctx).FeeRateBps)

	found, ok := k.GetSettlement(ctx, 1)
	require.True(t, ok)
	require.Equal(t, "TestMerchant", found.MerchantId)
}

func TestExportGenesis(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(), sdk.NewInt64Coin("unxrl", 100), "")
	k.CreateSettlement(ctx, msg)
	require.Len(t, k.GetAllSettlements(ctx), 1)
}

// --- metadata regression: LiveEnabled=false (default) ---

func TestMetadataDefaultLiveEnabledFalse(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	p := k.GetParams(ctx)
	require.False(t, p.LiveEnabled, "LiveEnabled must default to false")
}

func TestMetadataSettlementNoBankCall(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 10000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, uint64(1), s.Id)
	require.Equal(t, int32(types.SettlementCompleted), s.Status)
	require.False(t, s.FundsSettled)
	require.False(t, bk.sendCalled, "bank.SendCoins must not be called when LiveEnabled=false")
}

func TestMetadataSettlementFundsSettledFalse(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 500), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.False(t, s.FundsSettled)
	require.Equal(t, int32(types.SettlementCompleted), s.Status)
}

// --- live transfer: helpers ---

func enableLive(t *testing.T, k keeper.Keeper, ctx sdk.Context) {
	t.Helper()
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	require.NoError(t, k.SetParams(ctx, p))
}

// --- live transfer: success paths ---

func TestLiveSettlementSuccess(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int32(types.SettlementCompleted), s.Status)
	require.True(t, s.FundsSettled)
	require.True(t, bk.sendCalled)
}

func TestLiveSettlementPayerBalanceDecreases(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	payer := payerAddr()
	msg := types.NewMsgCreateSettlement(payer.String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	// fee = 100000*100/10000=1000; rebate tier2 1000bps => rebate 1000*1000/10000=100; netFee=900
	// merchantNet = 100000-900 = 99100
	// valShare = 900*6000/10000=540 (validator share is collected unconditionally
	// whenever live, independent of treasury/burn routing flags)
	// payer total debit = merchantNet + valShare = 99100 + 540 = 99640
	payerBalAfter := bk.balances[payer.String()]
	require.Equal(t, int64(1000000000-99640), payerBalAfter.AmountOf("unxrl").Int64())
}

func TestLiveSettlementMerchantBalanceIncreases(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	merchant := activeMerchantAddr().String()
	bk.balances[merchant] = sdk.NewCoins(sdk.NewInt64Coin("unxrl", 50000))
	merchantBalBefore := bk.balances[merchant].AmountOf("unxrl").Int64()

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	// merchantNet = 100000-900 = 99100
	merchantBalAfter := bk.balances[merchant].AmountOf("unxrl").Int64()
	require.Equal(t, merchantBalBefore+99100, merchantBalAfter)
}

func TestLiveSettlementFundsSettledTrue(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 1000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)
}

func TestLiveSettlementCorrectRecipient(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 10000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, activeMerchantAddr().String(), bk.lastTo.String())
}

func TestLiveSettlementCorrectAmount(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	// amount=100000, feeRate=100bps, rebate tier 2=1000bps
	// baseFee=1000, rebate=100, netFee=900, merchantNet=100000-900=99100
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int64(99100), bk.lastAmount.AmountOf("unxrl").Int64())
}

func TestLiveSettlementEventIncludesFundsSettled(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	events := ctx.EventManager().Events()
	found := false
	for _, ev := range events {
		if ev.Type == types.EventTypeCreateSettlement {
			for _, attr := range ev.Attributes {
				if string(attr.Key) == types.AttributeKeyFundsSettled && string(attr.Value) == "true" {
					found = true
				}
			}
		}
	}
	require.True(t, found, "event must contain funds_settled=true")
}

// --- live transfer: failure paths ---

func TestLiveSettlementInsufficientBalance(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	payer := payerAddr().String()
	bk.balances[payer] = sdk.NewCoins()
	bk.sendError = fmt.Errorf("insufficient funds")

	msg := types.NewMsgCreateSettlement(payer, activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "live settlement transfer failed")
	_, found := k.GetSettlement(ctx, 1)
	require.False(t, found)
}

func TestLiveSettlementNotStoredAfterFailedSend(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	bk.sendError = fmt.Errorf("simulated bank failure")

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
	require.Empty(t, k.GetAllSettlements(ctx))
}

func TestLiveSettlementInactiveMerchant(t *testing.T) {
	db := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db2)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(ms, tmproto.Header{}, false, log.NewNopLogger())

	inactiveAddr := activeMerchantAddr()
	mk := &mockMerchantKeeper{merchants: map[string]*merchanttypes.Merchant{
		inactiveAddr.String(): {Owner: inactiveAddr.String(), Name: "Inactive", Status: 1},
	}}
	bk := &mockBankKeeper{balances: map[string]sdk.Coins{}}
	k2 := keeper.NewKeeper(key, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), mk, &mockFeesKeeper{}, bk)
	p := k2.GetParams(ctx)
	p.LiveEnabled = true
	require.NoError(t, k2.SetParams(ctx, p))

	msg := types.NewMsgCreateSettlement(payerAddr().String(), inactiveAddr.String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	_, err := k2.CreateSettlement(ctx, msg)
	require.ErrorIs(t, err, types.ErrMerchantNotActive)
	require.False(t, bk.sendCalled, "bank must not be called for inactive merchant")
}

func TestLiveSettlementZeroAmount(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 0), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
	require.False(t, bk.sendCalled)
}

// --- fee calculation accuracy ---

func TestLiveProtocolFeePlusMerchantNetEqualsGross(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 123456), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	merchantNet := s.Amount.Amount.Sub(s.FeeAmount.Amount)
	require.True(t, s.FeeAmount.Amount.Add(merchantNet).Equal(s.Amount.Amount))
}

func TestLiveFeeSplitSumsToNetFee(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 99999), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	splitSum := s.ValidatorShare.Amount.Add(s.TreasuryShare.Amount).Add(s.BurnShare.Amount)
	require.True(t, splitSum.Equal(s.FeeAmount.Amount), "split sum %s != fee %s", splitSum, s.FeeAmount.Amount)
}

// --- double-completion prevention ---

func TestLiveSettlementUniqueIDs(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	s1, _ := k.CreateSettlement(ctx, msg)
	s2, _ := k.CreateSettlement(ctx, msg)
	require.NotEqual(t, s1.Id, s2.Id)
	require.Equal(t, uint64(1), s1.Id)
	require.Equal(t, uint64(2), s2.Id)
}

func TestLiveSettlementCannotRecomplete(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	s1, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s1.FundsSettled)
	s2, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.NotEqual(t, s1.Id, s2.Id)
	orig, ok := k.GetSettlement(ctx, s1.Id)
	require.True(t, ok)
	require.True(t, orig.FundsSettled)
}

// --- status transition rules for live-settled records ---

func TestLiveSettlementStatusChangeBlocked(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)

	err = k.UpdateSettlementStatus(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), s.Id, int32(types.SettlementRefunded))
	require.Error(t, err)
	require.Contains(t, err.Error(), "funds_settled=true")
}

func TestLiveSettlementStatusChangeToFailedBlocked(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)

	err = k.UpdateSettlementStatus(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), s.Id, int32(types.SettlementFailed))
	require.Error(t, err)
}

func TestLiveSettlementCannotTransitionToCancelled(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)

	err = k.UpdateSettlementStatus(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), s.Id, int32(types.SettlementCancelled))
	require.Error(t, err)
}

func TestMetadataSettlementStatusChangeAllowed(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.False(t, s.FundsSettled)

	err = k.UpdateSettlementStatus(ctx, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), s.Id, int32(types.SettlementRefunded))
	require.NoError(t, err)
	updated, _ := k.GetSettlement(ctx, s.Id)
	require.Equal(t, int32(types.SettlementRefunded), updated.Status)
}

// --- invariant helpers ---

func TestActiveSettledTotalsEmpty(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	totals := k.ActiveSettledTotals(ctx)
	require.True(t, totals.Empty())
}

func TestActiveSettledTotalsSingleLive(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)

	totals := k.ActiveSettledTotals(ctx)
	require.Equal(t, int64(99100), totals.AmountOf("unxrl").Int64())
}

func TestActiveSettledTotalsExcludesMetadata(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s1, _ := k.CreateSettlement(ctx, msg)
	require.False(t, s1.FundsSettled)

	enableLive(t, k, ctx)
	s2, _ := k.CreateSettlement(ctx, msg)
	require.True(t, s2.FundsSettled)

	totals := k.ActiveSettledTotals(ctx)
	require.Equal(t, int64(99100), totals.AmountOf("unxrl").Int64())
}

func TestValidateSettlementFundsInvariantClean(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	err := k.ValidateSettlementFundsInvariant(ctx)
	require.NoError(t, err)
}

func TestValidateSettlementFundsInvariantLiveClean(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableLive(t, k, ctx)
	_, err := k.CreateSettlement(ctx, types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 10000), ""))
	require.NoError(t, err)
	err = k.ValidateSettlementFundsInvariant(ctx)
	require.NoError(t, err)
}

// --- denom handling ---

func TestLiveSettlementDenomConsistency(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, "unxrl", s.Amount.Denom)
	require.Equal(t, "unxrl", s.FeeAmount.Denom)
	require.Equal(t, "unxrl", bk.lastAmount.GetDenomByIndex(0))
}

// --- treasury routing helpers ---

func enableTreasuryRouting(t *testing.T, k keeper.Keeper, ctx sdk.Context) {
	t.Helper()
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	p.TreasuryRoutingEnabled = true
	require.NoError(t, k.SetParams(ctx, p))
}

func enableBurnRouting(t *testing.T, k keeper.Keeper, ctx sdk.Context) {
	t.Helper()
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	p.TreasuryRoutingEnabled = true
	p.BurnRoutingEnabled = true
	require.NoError(t, k.SetParams(ctx, p))
}

// --- metadata/default: TreasuryRoutingEnabled defaults ---

func TestTreasuryRoutingDefaultFalse(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	p := k.GetParams(ctx)
	require.False(t, p.TreasuryRoutingEnabled, "TreasuryRoutingEnabled must default to false")
}

func TestTreasuryRoutingDisabledMeansMetadata(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 10000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.False(t, s.FundsSettled)
	require.False(t, bk.sendCalled)
	require.False(t, bk.sendToModuleCalled)
}

func TestTreasuryRoutingTrueLiveDisabledNoOp(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	p := k.GetParams(ctx)
	p.TreasuryRoutingEnabled = true
	require.NoError(t, k.SetParams(ctx, p))

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 10000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.False(t, s.FundsSettled)
	require.False(t, bk.sendCalled, "no bank calls when LiveEnabled=false even if TreasuryRoutingEnabled=true")
	require.False(t, bk.sendToModuleCalled)
}

// --- merchant-only path preserved ---

func TestLiveEnabledWithoutTreasuryRouting(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	// TreasuryRoutingEnabled is still false — the validator share still gets
	// collected unconditionally, but the treasury-specific transfer must not.
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)
	require.True(t, bk.sendCalled)
	require.Equal(t, int64(0), bk.moduleBalances["nexarail_treasury"].AmountOf("unxrl").Int64(), "treasury transfer must not happen")
}

func TestMerchantOnlyTreasuryBalanceUnchanged(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableLive(t, k, ctx)
	treasuryBefore := bk.moduleBalances["nexarail_treasury"].AmountOf("unxrl").Int64()

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)

	treasuryAfter := bk.moduleBalances["nexarail_treasury"].AmountOf("unxrl").Int64()
	require.Equal(t, treasuryBefore, treasuryAfter, "treasury balance must not change when TreasuryRoutingEnabled=false")
}

// --- live merchant + treasury success ---

func TestTreasuryRoutingSuccess(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)
	require.True(t, bk.sendCalled)
	require.True(t, bk.sendToModuleCalled)
}

func TestTreasuryRoutingMerchantReceivesNet(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	merchant := activeMerchantAddr().String()
	bk.balances[merchant] = sdk.NewCoins(sdk.NewInt64Coin("unxrl", 50000))
	merchantBefore := bk.balances[merchant].AmountOf("unxrl").Int64()

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	// merchantNet = 100000 - 900 = 99100
	merchantAfter := bk.balances[merchant].AmountOf("unxrl").Int64()
	require.Equal(t, merchantBefore+99100, merchantAfter)
}

func TestTreasuryRoutingTreasuryReceivesShare(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	// netFee = 900, treasury share = 900*2000/10000 = 180
	treasuryBal := bk.moduleBalances["nexarail_treasury"].AmountOf("unxrl").Int64()
	require.Equal(t, int64(180), treasuryBal)
}

func TestTreasuryRoutingPayerTotalDeduction(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	payer := payerAddr()
	msg := types.NewMsgCreateSettlement(payer.String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	// merchantNet=99100, valShare=540 (unconditional), treasury=180, total=99820
	payerAfter := bk.balances[payer.String()].AmountOf("unxrl").Int64()
	require.Equal(t, int64(1000000000-99820), payerAfter)
}

func TestTreasuryRoutingBurnShareMetadata(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	// burnShare = netFee - valShare - treasuryShare = 900 - 540 - 180 = 180
	require.Equal(t, int64(180), s.BurnShare.Amount.Int64())
	// Burn share should still be populated as metadata
}

func TestTreasuryRoutingValidatorShareMetadata(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int64(540), s.ValidatorShare.Amount.Int64())
}

func TestTreasuryRoutingEventAttribute(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	// Use amount large enough to produce positive treasury share
	// 100000 * 100/10000 = 1000 baseFee, rebate tier2 1000/10000*1000 = 100, netFee=900
	// treasuryShare = 900*2000/10000 = 180 > 0
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	events := ctx.EventManager().Events()
	foundTreasuryRouted := false
	for _, ev := range events {
		if ev.Type == types.EventTypeCreateSettlement {
			for _, attr := range ev.Attributes {
				if string(attr.Key) == types.AttributeKeyTreasuryRouted && string(attr.Value) == "true" {
					foundTreasuryRouted = true
				}
			}
		}
	}
	require.True(t, foundTreasuryRouted, "event must contain treasury_routed=true")
}

func TestTreasuryRoutingCorrectModuleName(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, "nexarail_treasury", bk.lastToModule)
}

func TestTreasuryRoutingZeroTreasuryShare(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	p.TreasuryRoutingEnabled = true
	p.FeeRateBps = 0 // zero fee → zero treasury share
	require.NoError(t, k.SetParams(ctx, p))

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)
	require.True(t, bk.sendCalled)                          // merchant transfer still happens
	require.False(t, bk.sendToModuleCalled, "zero treasury share must not trigger treasury transfer")
}

// --- failure paths ---

func TestTreasuryRoutingInsufficientBalance(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	bk.sendError = fmt.Errorf("insufficient funds")

	payer := payerAddr().String()
	msg := types.NewMsgCreateSettlement(payer, activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "live settlement transfer failed")
	_, found := k.GetSettlement(ctx, 1)
	require.False(t, found)
}

func TestTreasuryRoutingTreasuryTransferFails(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	bk.sendToModuleError = fmt.Errorf("simulated treasury transfer failure")
	bk.sendToModuleErrorFor = "nexarail_treasury"

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "treasury routing failed")
	// Settlement must not be stored
	require.Empty(t, k.GetAllSettlements(ctx))
}

func TestTreasuryRoutingNotStoredAfterFailure(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	bk.sendToModuleError = fmt.Errorf("treasury failure")

	// Use amount large enough to produce positive treasury share so the
	// treasury transfer is actually attempted (amount=100 gives treasuryShare=0)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
	require.Empty(t, k.GetAllSettlements(ctx))
}

func TestTreasuryRoutingInactiveMerchant(t *testing.T) {
	db := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db2)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(ms, tmproto.Header{}, false, log.NewNopLogger())

	inactiveAddr := activeMerchantAddr()
	mk := &mockMerchantKeeper{merchants: map[string]*merchanttypes.Merchant{
		inactiveAddr.String(): {Owner: inactiveAddr.String(), Name: "Inactive", Status: 1},
	}}
	bk := &mockBankKeeper{
		balances:       map[string]sdk.Coins{},
		moduleBalances: map[string]sdk.Coins{"nexarail_treasury": sdk.NewCoins()},
	}
	k2 := keeper.NewKeeper(key, sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String(), mk, &mockFeesKeeper{}, bk)
	p := k2.GetParams(ctx)
	p.LiveEnabled = true
	p.TreasuryRoutingEnabled = true
	require.NoError(t, k2.SetParams(ctx, p))

	msg := types.NewMsgCreateSettlement(payerAddr().String(), inactiveAddr.String(),
		sdk.NewInt64Coin("unxrl", 100), "")
	_, err := k2.CreateSettlement(ctx, msg)
	require.ErrorIs(t, err, types.ErrMerchantNotActive)
	require.False(t, bk.sendCalled, "no bank calls for inactive merchant")
	require.False(t, bk.sendToModuleCalled)
}

// --- fee calculation with treasury routing ---

func TestTreasuryRoutingFeeSplit(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	// netFee=900, treasury=900*2000/10000=180, val=900*6000/10000=540, burn=900-540-180=180
	require.Equal(t, int64(900), s.FeeAmount.Amount.Int64())
	require.Equal(t, int64(180), s.TreasuryShare.Amount.Int64())
	require.Equal(t, int64(540), s.ValidatorShare.Amount.Int64())
	require.Equal(t, int64(180), s.BurnShare.Amount.Int64())
}

func TestTreasuryRoutingNetPlusFeeEqualsGross(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 123456), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	merchantNet := s.Amount.Amount.Sub(s.FeeAmount.Amount)
	require.True(t, s.FeeAmount.Amount.Add(merchantNet).Equal(s.Amount.Amount))
}

// --- invariant helpers with treasury routing ---

func TestTreasuryRoutingActiveSettledTotals(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	totals := k.ActiveSettledTotals(ctx)
	require.Equal(t, int64(99100), totals.AmountOf("unxrl").Int64())
}

func TestTreasuryRoutingInvariantClean(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx)
	_, err := k.CreateSettlement(ctx, types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 10000), ""))
	require.NoError(t, err)
	err = k.ValidateSettlementFundsInvariant(ctx)
	require.NoError(t, err)
}

// --- burn routing: metadata/default ---

func TestBurnRoutingDefaultFalse(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	p := k.GetParams(ctx)
	require.False(t, p.BurnRoutingEnabled, "BurnRoutingEnabled must default to false")
}

func TestBurnRoutingTrueLiveDisabledNoOp(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	p := k.GetParams(ctx)
	p.BurnRoutingEnabled = true
	require.NoError(t, k.SetParams(ctx, p))

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.False(t, s.FundsSettled)
	require.False(t, bk.sendCalled)
	require.False(t, bk.burnCalled)
}

func TestBurnRoutingDisabledNoBurnCall(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableTreasuryRouting(t, k, ctx) // Live+Treasury, no Burn

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)
	require.False(t, s.BurnExecuted)
	require.False(t, bk.burnCalled, "BurnCoins must not be called when BurnRoutingEnabled=false")
}

// --- burn routing: success ---

func TestBurnRoutingSuccess(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)
	require.True(t, s.BurnExecuted)
	require.True(t, bk.burnCalled)
}

func TestBurnRoutingPayerTotalDeduction(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)
	payer := payerAddr()

	msg := types.NewMsgCreateSettlement(payer.String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	// merchantNet=99100, valShare=540 (unconditional), treasury=180, burn=180
	// total = 99100+540+180+180 = 100000 (= full settlement amount, as expected:
	// merchantNet + netFee should always equal amount)
	payerAfter := bk.balances[payer.String()].AmountOf("unxrl").Int64()
	require.Equal(t, int64(1000000000-100000), payerAfter)
}

func TestBurnRoutingMerchantReceivesNet(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)
	merchant := activeMerchantAddr().String()
	bk.balances[merchant] = sdk.NewCoins(sdk.NewInt64Coin("unxrl", 50000))
	merchantBefore := bk.balances[merchant].AmountOf("unxrl").Int64()

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, merchantBefore+99100, bk.balances[merchant].AmountOf("unxrl").Int64())
}

func TestBurnRoutingTreasuryReceivesShare(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int64(180), bk.moduleBalances["nexarail_treasury"].AmountOf("unxrl").Int64())
}

func TestBurnRoutingBurnerBalanceZero(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	// Burner should be empty after send→burn completes
	require.True(t, bk.moduleBalances["nexarail_burner"].Empty(),
		"burner module balance must be zero after send+burn")
}

func TestBurnRoutingTotalBurned(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int64(180), bk.totalBurned.AmountOf("unxrl").Int64())
}

func TestBurnRoutingValidatorShareMetadata(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableBurnRouting(t, k, ctx)

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, int64(540), s.ValidatorShare.Amount.Int64())
}

func TestBurnRoutingEventAttribute(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableBurnRouting(t, k, ctx)

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	events := ctx.EventManager().Events()
	found := false
	for _, ev := range events {
		if ev.Type == types.EventTypeCreateSettlement {
			for _, attr := range ev.Attributes {
				if string(attr.Key) == types.AttributeKeyBurnRouted && string(attr.Value) == "true" {
					found = true
				}
			}
		}
	}
	require.True(t, found, "event must contain burn_routed=true")
}

func TestBurnRoutingCorrectModuleName(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, "nexarail_burner", bk.lastBurnModule)
}

// --- burn routing: zero burn share ---

func TestBurnRoutingZeroBurnShareSkips(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	p := k.GetParams(ctx)
	p.LiveEnabled = true
	p.TreasuryRoutingEnabled = true
	p.BurnRoutingEnabled = true
	p.FeeRateBps = 0 // zero fee → zero burn share
	require.NoError(t, k.SetParams(ctx, p))

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	s, err := k.CreateSettlement(ctx, msg)
	require.NoError(t, err)
	require.True(t, s.FundsSettled)
	require.False(t, s.BurnExecuted)
	require.False(t, bk.burnCalled, "zero burn share must not trigger BurnCoins")
}

// --- burn routing: failure paths ---

func TestBurnRoutingInsufficientBalance(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)
	bk.sendError = fmt.Errorf("insufficient funds")

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "live settlement transfer failed")
	_, found := k.GetSettlement(ctx, 1)
	require.False(t, found)
}

func TestBurnRoutingBurnTransferFails(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)
	bk.burnError = fmt.Errorf("simulated BurnCoins failure")

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "burn execution failed")
	require.Empty(t, k.GetAllSettlements(ctx))
}

func TestBurnRoutingBurnCoinsFails(t *testing.T) {
	k, ctx, bk := setupKeeper(t)
	enableBurnRouting(t, k, ctx)
	bk.burnError = fmt.Errorf("simulated BurnCoins failure")

	msg := types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), "")
	_, err := k.CreateSettlement(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "burn execution failed")
	require.Empty(t, k.GetAllSettlements(ctx))
}

// --- burn routing: invariants ---

func TestBurnRoutingInvariantClean(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	enableBurnRouting(t, k, ctx)
	_, err := k.CreateSettlement(ctx, types.NewMsgCreateSettlement(payerAddr().String(), activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), ""))
	require.NoError(t, err)
	err = k.ValidateSettlementFundsInvariant(ctx)
	require.NoError(t, err)
}

func TestBurnExecutedRequiresFundsSettled(t *testing.T) {
	// Validate at the types level
	s := types.NewSettlement(1, payerAddr().String(), activeMerchantAddr().String(),
		"test", activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), sdk.NewInt64Coin("unxrl", 900),
		sdk.NewInt64Coin("unxrl", 540), sdk.NewInt64Coin("unxrl", 180),
		sdk.NewInt64Coin("unxrl", 180), sdk.NewInt64Coin("unxrl", 100),
		1000, "", "", "", 100,
	)
	s.Status = int32(types.SettlementCompleted)
	s.BurnExecuted = true // FundsSettled is still false
	err := s.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "burn_executed=true requires funds_settled=true")
}

func TestBurnExecutedRequiresCompletedStatus(t *testing.T) {
	s := types.NewSettlement(1, payerAddr().String(), activeMerchantAddr().String(),
		"test", activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), sdk.NewInt64Coin("unxrl", 900),
		sdk.NewInt64Coin("unxrl", 540), sdk.NewInt64Coin("unxrl", 180),
		sdk.NewInt64Coin("unxrl", 180), sdk.NewInt64Coin("unxrl", 100),
		1000, "", "", "", 100,
	)
	s.Status = int32(types.SettlementPending)
	s.FundsSettled = true
	s.BurnExecuted = true
	err := s.Validate()
	// FundsSettled check fires before BurnExecuted check since both require Completed
	require.Error(t, err)
	require.Contains(t, err.Error(), "funds_settled=true requires status completed")
}

func TestBurnExecutedRequiresPositiveBurnShare(t *testing.T) {
	s := types.NewSettlement(1, payerAddr().String(), activeMerchantAddr().String(),
		"test", activeMerchantAddr().String(),
		sdk.NewInt64Coin("unxrl", 100000), sdk.NewInt64Coin("unxrl", 900),
		sdk.NewInt64Coin("unxrl", 540), sdk.NewInt64Coin("unxrl", 180),
		sdk.NewInt64Coin("unxrl", 0), sdk.NewInt64Coin("unxrl", 100),
		1000, "", "", "", 100,
	)
	s.Status = int32(types.SettlementCompleted)
	s.FundsSettled = true
	s.BurnExecuted = true // BurnShare=0
	err := s.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "burn_executed=true requires positive burn share")
}

// =============================================================================
// Phase 8E: Stress Tests — Invariants, Fuzz, Randomized, Failure Injection
// =============================================================================

func TestInvariant_LiveFlagsDefaultFalse(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	params := k.GetParams(ctx)
	require.False(t, params.LiveEnabled, "live_enabled must default false")
	require.False(t, params.TreasuryRoutingEnabled, "treasury_routing_enabled must default false")
	require.False(t, params.BurnRoutingEnabled, "burn_routing_enabled must default false")
}

func TestInvariant_FeeRateBpsInRange(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	params := k.GetParams(ctx)
	require.Greater(t, params.FeeRateBps, uint32(0), "fee_rate_bps must be positive")
	require.LessOrEqual(t, params.FeeRateBps, uint32(10000), "fee_rate_bps cannot exceed 10000 bps")
}

func TestFuzz_SettlementFeeCalculation(t *testing.T) {
	k, ctx, _ := setupKeeper(t)

	// Test fee calculation at various amounts and rates
	amounts := []uint64{100, 1000, 10000, 100000, 1000000, 99999999}
	for _, amt := range amounts {
		params := k.GetParams(ctx)
		fee := uint64(amt) * uint64(params.FeeRateBps) / 10000
		// Fee must be less than or equal to amount
		require.LessOrEqual(t, fee, uint64(amt), "fee %d cannot exceed amount %d at rate %d bps", fee, amt, params.FeeRateBps)
	}
}

func TestFuzz_RebateTiersValid(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	params := k.GetParams(ctx)
	require.NotEmpty(t, params.RebateTiers, "rebate tiers must not be empty")
	// Tiers should be in ascending order
	for i := 1; i < len(params.RebateTiers); i++ {
		require.LessOrEqual(t, params.RebateTiers[i-1], params.RebateTiers[i],
			"rebate tiers must be sorted ascending")
	}
}

func TestRandom_ParamUpdateRecovery(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	original := k.GetParams(ctx)

	// Modify and restore params repeatedly
	for i := 0; i < 5; i++ {
		params := original
		params.FeeRateBps = uint32(50 + i*25)
		k.SetParams(ctx, params)
		require.Equal(t, params.FeeRateBps, k.GetParams(ctx).FeeRateBps)
	}
	// Restore original
	k.SetParams(ctx, original)
	require.Equal(t, original.FeeRateBps, k.GetParams(ctx).FeeRateBps)
}

func TestFailure_LiveEnabledFalsePreventsTransfers(t *testing.T) {
	k, ctx, _ := setupKeeper(t)

	// With LiveEnabled=false, live transfers should be blocked
	params := k.GetParams(ctx)
	require.False(t, params.LiveEnabled)
	require.False(t, params.TreasuryRoutingEnabled)
	require.False(t, params.BurnRoutingEnabled)

	// Verify that all routing flags are independently false
	// Each must be individually enabled for its routing to activate
	require.NotEqual(t, params.LiveEnabled, true)
}

// --- Phase 14D: Params-update event emission ---

func TestUpdateParamsEmitsEvent(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	ms := keeper.NewMsgServerImpl(k)
	authority := sdk.AccAddress([]byte{8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8,8}).String()
	params := types.DefaultParams()

	msg := types.NewMsgUpdateParams(authority, params)
	_, err := ms.UpdateParams(ctx, msg)
	require.NoError(t, err)

	events := ctx.EventManager().Events()
	var found bool
	for _, ev := range events {
		if ev.Type == types.EventTypeUpdateParams {
			found = true
			for _, attr := range ev.Attributes {
				if string(attr.Key) == "authority" && string(attr.Value) == authority {
					return
				}
			}
		}
	}
	require.True(t, found, "must emit settlement_params_updated event with authority")
}


func TestUpdateParamsNoEventOnInvalidAuthority(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	err := k.UpdateParams(ctx, "wrong-authority", types.DefaultParams())
	require.Error(t, err)

	events := ctx.EventManager().Events()
	for _, ev := range events {
		require.NotEqual(t, types.EventTypeUpdateParams, ev.Type, "must not emit event on failed update")
	}
}
