package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/nexarail/chain/x/common"
	feessdk "github.com/nexarail/chain/x/fees"
	feeskeeper "github.com/nexarail/chain/x/fees/keeper"
	feestypes "github.com/nexarail/chain/x/fees/types"

	merchantsdk "github.com/nexarail/chain/x/merchant"
	merchantkeeper "github.com/nexarail/chain/x/merchant/keeper"
	merchanttypes "github.com/nexarail/chain/x/merchant/types"

	settlementsdk "github.com/nexarail/chain/x/settlement"
	settlementkeeper "github.com/nexarail/chain/x/settlement/keeper"
	settlementtypes "github.com/nexarail/chain/x/settlement/types"

	escrowsdk "github.com/nexarail/chain/x/escrow"
	escrowkeeper "github.com/nexarail/chain/x/escrow/keeper"
	escrowtypes "github.com/nexarail/chain/x/escrow/types"

	payoutsdk "github.com/nexarail/chain/x/payout"
	payoutkeeper "github.com/nexarail/chain/x/payout/keeper"
	payouttypes "github.com/nexarail/chain/x/payout/types"

	treasurysdk "github.com/nexarail/chain/x/treasury"
	treasurykeeper "github.com/nexarail/chain/x/treasury/keeper"
	treasurytypes "github.com/nexarail/chain/x/treasury/types"
)

const (
	AccountAddressPrefix = "nxr"
	Name                 = "nexarail"

	// Module account names for Phase 5B live fund infrastructure.
	// Treasury and Burner alias x/common's canonical constants — see
	// x/common/accounts.go — since those two names are also referenced
	// independently by x/treasury, x/payout, and x/settlement.
	NexaRailEscrowModuleAccount    = "nexarail_escrow"
	NexaRailTreasuryModuleAccount  = common.TreasuryModuleAccount
	NexaRailFeeRouterModuleAccount = "nexarail_fee_router"
	NexaRailBurnerModuleAccount    = common.BurnerModuleAccount
)

var DefaultNodeHome = filepath.Join(os.Getenv("HOME"), ".nexarail")

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distr.AppModuleBasic{},
	gov.NewAppModuleBasic(nil),
	params.AppModuleBasic{},
	crisis.AppModuleBasic{},
	slashing.AppModuleBasic{},
	vesting.AppModuleBasic{},
	feegrantmodule.AppModuleBasic{},
	authzmodule.AppModuleBasic{},
	evidence.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	feessdk.AppModuleBasic{},
	merchantsdk.AppModuleBasic{},
	settlementsdk.AppModuleBasic{}, escrowsdk.AppModuleBasic{}, treasurysdk.AppModuleBasic{}, payoutsdk.AppModuleBasic{},
)

var _ servertypes.Application = (*NexaRailApp)(nil)

type NexaRailApp struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    *stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        *govkeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
	EvidenceKeeper   evidencekeeper.Keeper
	FeeGrantKeeper   feegrantkeeper.Keeper
	AuthzKeeper      authzkeeper.Keeper
	FeesKeeper       feeskeeper.Keeper
	MerchantKeeper   merchantkeeper.Keeper
	SettlementKeeper settlementkeeper.Keeper
	EscrowKeeper     escrowkeeper.Keeper
	PayoutKeeper     payoutkeeper.Keeper
	TreasuryKeeper   treasurykeeper.Keeper

	mm *module.Manager
	sm *module.SimulationManager
}

func NewNexaRailApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *NexaRailApp {
	appCodec := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	SetBech32Prefix()

	bApp := baseapp.NewBaseApp(Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	// Set up a small ParamStore for consensus params. BaseApp calls this during
	// PrepareProposal/ProcessProposal, including after process restarts.
	bApp.SetParamStore(newParamStore(homePath))

	app := &NexaRailApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
	}

	// Store keys
	app.keys = sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, capabilitytypes.StoreKey, feegrant.StoreKey,
		authz.ModuleName, crisistypes.StoreKey, feestypes.StoreKey,
		merchanttypes.StoreKey, settlementtypes.StoreKey, payouttypes.StoreKey, escrowtypes.StoreKey,
		treasurytypes.StoreKey,
	)
	app.tkeys = sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	app.memKeys = sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		crisistypes.ModuleName:         nil,
		merchanttypes.ModuleName:       nil,
		NexaRailEscrowModuleAccount:    nil,
		NexaRailTreasuryModuleAccount:  nil,
		NexaRailFeeRouterModuleAccount: nil,
		NexaRailBurnerModuleAccount:    {authtypes.Burner},
	}

	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// Keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, app.keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount, maccPerms,
		AccountAddressPrefix, authority,
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec, app.keys[banktypes.StoreKey],
		app.AccountKeeper, blockedAddrs, authority,
	)
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec, app.keys[capabilitytypes.StoreKey],
		app.memKeys[capabilitytypes.MemStoreKey],
	)
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec, app.keys[stakingtypes.StoreKey],
		app.AccountKeeper, app.BankKeeper, authority,
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec, app.keys[minttypes.StoreKey],
		app.StakingKeeper, app.AccountKeeper, app.BankKeeper,
		authtypes.FeeCollectorName, authority,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, app.keys[distrtypes.StoreKey],
		app.AccountKeeper, app.BankKeeper, app.StakingKeeper,
		authtypes.FeeCollectorName, authority,
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, legacyAmino, app.keys[slashingtypes.StoreKey],
		app.StakingKeeper, authority,
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec, app.keys[crisistypes.StoreKey],
		invCheckPeriod, app.BankKeeper,
		authtypes.FeeCollectorName, authority,
	)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights, app.keys[upgradetypes.StoreKey],
		appCodec, homePath, app.BaseApp, authority,
	)
	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec, app.keys[feegrant.StoreKey], app.AccountKeeper,
	)
	app.AuthzKeeper = authzkeeper.NewKeeper(
		app.keys[authz.ModuleName], appCodec,
		app.MsgServiceRouter(), app.AccountKeeper,
	)
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, app.keys[evidencetypes.StoreKey],
		app.StakingKeeper, app.SlashingKeeper,
	)
	app.EvidenceKeeper = *evidenceKeeper

	govConfig := govtypes.DefaultConfig()
	app.GovKeeper = govkeeper.NewKeeper(
		appCodec, app.keys[govtypes.StoreKey],
		app.AccountKeeper, app.BankKeeper, app.StakingKeeper,
		app.MsgServiceRouter(), govConfig, authority,
	)
	app.ParamsKeeper = initParamsKeeper(
		appCodec, legacyAmino,
		app.keys[paramstypes.StoreKey], app.tkeys[paramstypes.TStoreKey],
	)

	// Fees keeper
	app.FeesKeeper = feeskeeper.NewKeeper(
		app.keys[feestypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		authority,
	)

	// Merchant keeper
	app.MerchantKeeper = merchantkeeper.NewKeeper(
		app.keys[merchanttypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		authority,
	)

	// Settlement keeper
	app.SettlementKeeper = settlementkeeper.NewKeeper(
		app.keys[settlementtypes.StoreKey],
		authority,
		app.MerchantKeeper,
		app.FeesKeeper,
		app.BankKeeper,
	)

	// Escrow keeper
	app.EscrowKeeper = escrowkeeper.NewKeeper(
		app.keys[escrowtypes.StoreKey],
		authority,
		app.MerchantKeeper,
		app.BankKeeper,
	)

	// Payout keeper
	app.PayoutKeeper = payoutkeeper.NewKeeper(
		app.keys[payouttypes.StoreKey],
		authority,
		app.MerchantKeeper,
		app.BankKeeper,
	)

	// Treasury keeper
	app.TreasuryKeeper = treasurykeeper.NewKeeper(
		app.keys[treasurytypes.StoreKey],
		authority,
		app.BankKeeper,
	)

	// Register staking hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
		),
	)

	// Subspaces for legacy modules (already registered in initParamsKeeper)
	authSubspace, _ := app.ParamsKeeper.GetSubspace(authtypes.ModuleName)
	bankSubspace, _ := app.ParamsKeeper.GetSubspace(banktypes.ModuleName)
	stakingSubspace, _ := app.ParamsKeeper.GetSubspace(stakingtypes.ModuleName)
	mintSubspace, _ := app.ParamsKeeper.GetSubspace(minttypes.ModuleName)
	distrSubspace, _ := app.ParamsKeeper.GetSubspace(distrtypes.ModuleName)
	slashingSubspace, _ := app.ParamsKeeper.GetSubspace(slashingtypes.ModuleName)
	govSubspace, _ := app.ParamsKeeper.GetSubspace(govtypes.ModuleName)
	crisisSubspace, _ := app.ParamsKeeper.GetSubspace(crisistypes.ModuleName)

	// Module manager
	app.mm = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx, encodingConfig.TxConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, authSubspace),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, bankSubspace),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(app.CrisisKeeper, false, crisisSubspace),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, govSubspace),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, mintSubspace),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, slashingSubspace),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, distrSubspace),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, stakingSubspace),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		feessdk.NewAppModule(app.FeesKeeper),
		merchantsdk.NewAppModule(app.MerchantKeeper),
		settlementsdk.NewAppModule(app.SettlementKeeper), escrowsdk.NewAppModule(app.EscrowKeeper), treasurysdk.NewAppModule(app.TreasuryKeeper), payoutsdk.NewAppModule(app.PayoutKeeper),
	)

	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName, authtypes.ModuleName, vestingtypes.ModuleName,
		banktypes.ModuleName, distrtypes.ModuleName, stakingtypes.ModuleName,
		slashingtypes.ModuleName, govtypes.ModuleName, minttypes.ModuleName,
		crisistypes.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName,
		feegrant.ModuleName, authz.ModuleName, paramstypes.ModuleName, upgradetypes.ModuleName,
		feestypes.ModuleName, merchanttypes.ModuleName, settlementtypes.ModuleName, escrowtypes.ModuleName, treasurytypes.ModuleName, payouttypes.ModuleName,
	)
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName, capabilitytypes.ModuleName,
		minttypes.ModuleName, distrtypes.ModuleName,
		slashingtypes.ModuleName, evidencetypes.ModuleName, stakingtypes.ModuleName,
		authtypes.ModuleName, authz.ModuleName, banktypes.ModuleName,
		crisistypes.ModuleName, feegrant.ModuleName, genutiltypes.ModuleName,
		govtypes.ModuleName, paramstypes.ModuleName, vestingtypes.ModuleName,
		feestypes.ModuleName, merchanttypes.ModuleName, settlementtypes.ModuleName, escrowtypes.ModuleName, treasurytypes.ModuleName, payouttypes.ModuleName,
	)
	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName, govtypes.ModuleName, stakingtypes.ModuleName,
		authtypes.ModuleName, authz.ModuleName, banktypes.ModuleName,
		capabilitytypes.ModuleName, distrtypes.ModuleName, evidencetypes.ModuleName,
		feegrant.ModuleName, genutiltypes.ModuleName, minttypes.ModuleName,
		paramstypes.ModuleName, slashingtypes.ModuleName, upgradetypes.ModuleName,
		vestingtypes.ModuleName, feestypes.ModuleName, merchanttypes.ModuleName, settlementtypes.ModuleName, escrowtypes.ModuleName, treasurytypes.ModuleName, payouttypes.ModuleName,
	)

	configurator := module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(configurator)

	// Phase 8F: Register no-op upgrade handlers for testnet
	// These prove upgrade infrastructure works without changing state.
	// WARNING: Do not enable on mainnet without full audit.
	app.registerUpgradeHandlers()

	// Ante handler
	anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
		FeegrantKeeper:  app.FeeGrantKeeper,
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})
	if err != nil {
		panic(err)
	}
	app.SetAnteHandler(anteHandler)

	app.MountKVStores(app.keys)
	app.MountTransientStores(app.tkeys)
	app.MountMemoryStores(app.memKeys)
	app.SetQueryMultiStore(newLatestQueryMultiStore(app.CommitMultiStore()))

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(err)
		}
	}

	// Simulation manager (minimal for devnet)
	app.sm = module.NewSimulationManager()
	app.sm.RegisterStoreDecoders()

	return app
}

// ABCI methods
func (app *NexaRailApp) Name() string { return app.BaseApp.Name() }

func (app *NexaRailApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

func (app *NexaRailApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

func (app *NexaRailApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// Utility methods
func (app *NexaRailApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func (app *NexaRailApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}
	return modAccAddrs
}

func (app *NexaRailApp) LegacyAmino() *codec.LegacyAmino              { return app.legacyAmino }
func (app *NexaRailApp) AppCodec() codec.Codec                        { return app.appCodec }
func (app *NexaRailApp) InterfaceRegistry() types.InterfaceRegistry   { return app.interfaceRegistry }
func (app *NexaRailApp) TxConfig() client.TxConfig                    { return app.txConfig }
func (app *NexaRailApp) SimulationManager() *module.SimulationManager { return app.sm }

// servertypes.Application interface
func (app *NexaRailApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	app.RegisterRuntimeReadbackRoutes(apiSvr.GRPCGatewayRouter)
	// Standard SDK gateway routes: /cosmos/tx/v1beta1/*, /cosmos/base/tendermint/v1beta1/*,
	// /cosmos/base/node/v1beta1/*. Explorers, wallets and the chain-registry CI depend on these.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

func (app *NexaRailApp) RegisterRuntimeReadbackRoutes(mux *runtime.ServeMux) {
	paramsContext := func() sdk.Context {
		return app.BaseApp.NewUncachedContext(false, tmproto.Header{})
	}

	common.RegisterQueryRoute(mux, "GET", "/nexarail/fees/v1/params", func() (interface{}, error) {
		return map[string]interface{}{"params": app.FeesKeeper.GetParams(paramsContext())}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/merchant/v1/params", func() (interface{}, error) {
		return map[string]interface{}{"params": app.MerchantKeeper.GetParams(paramsContext())}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/merchant/v1/merchants", func() (interface{}, error) {
		return map[string]interface{}{"merchants": app.MerchantKeeper.GetAllMerchants(paramsContext())}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/merchant/v1/merchant/{owner}", "owner", func(owner string) (interface{}, error) {
		if owner == "" {
			return nil, fmt.Errorf("merchant owner address required")
		}
		addr, err := sdk.AccAddressFromBech32(owner)
		if err != nil {
			return nil, fmt.Errorf("invalid merchant owner address '%s': %w", owner, err)
		}
		merchant, found := app.MerchantKeeper.GetMerchant(paramsContext(), addr)
		if !found {
			return nil, fmt.Errorf("merchant %s not found", owner)
		}
		return map[string]interface{}{"merchant": merchant}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/settlement/v1/params", func() (interface{}, error) {
		return map[string]interface{}{"params": app.SettlementKeeper.GetParams(paramsContext())}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/settlement/v1/settlements", func() (interface{}, error) {
		return map[string]interface{}{"settlements": app.SettlementKeeper.GetAllSettlements(paramsContext())}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/settlement/v1/settlement/{id}", "id", func(idRaw string) (interface{}, error) {
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("settlement id: %w", err)
		}
		settlement, found := app.SettlementKeeper.GetSettlement(paramsContext(), id)
		if !found {
			return nil, fmt.Errorf("settlement %d not found", id)
		}
		return map[string]interface{}{"settlement": settlement}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/settlement/v1/settlements/by-merchant/{owner}", "owner", func(owner string) (interface{}, error) {
		if owner == "" {
			return nil, fmt.Errorf("merchant owner address required")
		}
		return map[string]interface{}{"settlements": app.SettlementKeeper.GetSettlementsByMerchant(paramsContext(), owner)}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/settlement/v1/settlements/by-payer/{payer}", "payer", func(payer string) (interface{}, error) {
		if payer == "" {
			return nil, fmt.Errorf("payer address required")
		}
		return map[string]interface{}{"settlements": app.SettlementKeeper.GetSettlementsByPayer(paramsContext(), payer)}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/escrow/v1/params", func() (interface{}, error) {
		return map[string]interface{}{"params": app.EscrowKeeper.GetParams(paramsContext())}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/escrow/v1/escrows", func() (interface{}, error) {
		return map[string]interface{}{"escrows": app.EscrowKeeper.GetAllEscrows(paramsContext())}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/escrow/v1/escrow/{id}", "id", func(idRaw string) (interface{}, error) {
		if idRaw == "" {
			return nil, fmt.Errorf("escrow id required")
		}
		escrow, found := app.EscrowKeeper.GetEscrow(paramsContext(), idRaw)
		if !found {
			return nil, fmt.Errorf("escrow %s not found", idRaw)
		}
		return map[string]interface{}{"escrow": escrow}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/escrow/v1/escrows/by-buyer/{buyer}", "buyer", func(buyer string) (interface{}, error) {
		if buyer == "" {
			return nil, fmt.Errorf("buyer address required")
		}
		return map[string]interface{}{"escrows": app.EscrowKeeper.GetEscrowsByBuyer(paramsContext(), buyer)}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/escrow/v1/escrows/by-seller/{seller}", "seller", func(seller string) (interface{}, error) {
		if seller == "" {
			return nil, fmt.Errorf("seller address required")
		}
		return map[string]interface{}{"escrows": app.EscrowKeeper.GetEscrowsBySeller(paramsContext(), seller)}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/escrow/v1/escrows/by-merchant/{merchant}", "merchant", func(merchant string) (interface{}, error) {
		if merchant == "" {
			return nil, fmt.Errorf("merchant id required")
		}
		return map[string]interface{}{"escrows": app.EscrowKeeper.GetEscrowsByMerchant(paramsContext(), merchant)}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/escrow/v1/escrow/exists/{id}", "id", func(idRaw string) (interface{}, error) {
		if idRaw == "" {
			return nil, fmt.Errorf("escrow id required")
		}
		return map[string]interface{}{"exists": app.EscrowKeeper.HasEscrow(paramsContext(), idRaw)}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/payout/v1/params", func() (interface{}, error) {
		return map[string]interface{}{"params": app.PayoutKeeper.GetParams(paramsContext())}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/payout/v1/payouts", func() (interface{}, error) {
		return map[string]interface{}{"payouts": app.PayoutKeeper.GetAllPayouts(paramsContext())}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/payout/v1/payout/{id}", "id", func(id string) (interface{}, error) {
		if id == "" {
			return nil, fmt.Errorf("payout id required")
		}
		payout, found := app.PayoutKeeper.GetPayout(paramsContext(), id)
		if !found {
			return nil, fmt.Errorf("payout %s not found", id)
		}
		return map[string]interface{}{"payout": payout}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/payout/v1/payout/exists/{id}", "id", func(id string) (interface{}, error) {
		if id == "" {
			return nil, fmt.Errorf("payout id required")
		}
		return map[string]interface{}{"exists": app.PayoutKeeper.HasPayout(paramsContext(), id)}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/payout/v1/payouts/by-merchant/{merchant}", "merchant", func(merchant string) (interface{}, error) {
		if merchant == "" {
			return nil, fmt.Errorf("merchant id required")
		}
		return map[string]interface{}{"payouts": app.PayoutKeeper.GetPayoutsByMerchant(paramsContext(), merchant)}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/payout/v1/payouts/by-recipient/{recipient}", "recipient", func(recipient string) (interface{}, error) {
		if recipient == "" {
			return nil, fmt.Errorf("recipient address required")
		}
		return map[string]interface{}{"payouts": app.PayoutKeeper.GetPayoutsByRecipient(paramsContext(), recipient)}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/payout/v1/payouts/by-initiator/{initiator}", "initiator", func(initiator string) (interface{}, error) {
		if initiator == "" {
			return nil, fmt.Errorf("initiator address required")
		}
		return map[string]interface{}{"payouts": app.PayoutKeeper.GetPayoutsByInitiator(paramsContext(), initiator)}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/payout/v1/batch-payout/{id}", "id", func(id string) (interface{}, error) {
		if id == "" {
			return nil, fmt.Errorf("batch payout id required")
		}
		batch, found := app.PayoutKeeper.GetBatchPayout(paramsContext(), id)
		if !found {
			return nil, fmt.Errorf("batch payout %s not found", id)
		}
		return map[string]interface{}{"batch_payout": batch}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/payout/v1/batch-payouts", func() (interface{}, error) {
		return map[string]interface{}{"batch_payouts": app.PayoutKeeper.GetAllBatchPayouts(paramsContext())}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/treasury/v1/params", func() (interface{}, error) {
		return map[string]interface{}{"params": app.TreasuryKeeper.GetParams(paramsContext())}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/treasury/v1/summary", func() (interface{}, error) {
		ctx := paramsContext()
		return map[string]interface{}{
			"total_accounts":       len(app.TreasuryKeeper.GetAllTreasuryAccounts(ctx)),
			"total_budgets":        len(app.TreasuryKeeper.GetAllBudgets(ctx)),
			"total_grants":         len(app.TreasuryKeeper.GetAllGrants(ctx)),
			"total_spend_requests": len(app.TreasuryKeeper.GetAllSpendRequests(ctx)),
		}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/treasury/v1/spend/{id}", "id", func(id string) (interface{}, error) {
		if id == "" {
			return nil, fmt.Errorf("spend id required")
		}
		spend, found := app.TreasuryKeeper.GetSpendRequest(paramsContext(), id)
		if !found {
			return nil, fmt.Errorf("spend request %s not found", id)
		}
		return map[string]interface{}{"spend_request": spend}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/treasury/v1/spends", func() (interface{}, error) {
		return map[string]interface{}{"spend_requests": app.TreasuryKeeper.GetAllSpendRequests(paramsContext())}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/treasury/v1/account/{id}", "id", func(id string) (interface{}, error) {
		if id == "" {
			return nil, fmt.Errorf("account id required")
		}
		account, found := app.TreasuryKeeper.GetTreasuryAccount(paramsContext(), id)
		if !found {
			return nil, fmt.Errorf("treasury account %s not found", id)
		}
		return map[string]interface{}{"treasury_account": account}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/treasury/v1/accounts", func() (interface{}, error) {
		return map[string]interface{}{"treasury_accounts": app.TreasuryKeeper.GetAllTreasuryAccounts(paramsContext())}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/treasury/v1/budget/{id}", "id", func(id string) (interface{}, error) {
		if id == "" {
			return nil, fmt.Errorf("budget id required")
		}
		budget, found := app.TreasuryKeeper.GetBudget(paramsContext(), id)
		if !found {
			return nil, fmt.Errorf("budget %s not found", id)
		}
		return map[string]interface{}{"budget": budget}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/treasury/v1/budgets", func() (interface{}, error) {
		return map[string]interface{}{"budgets": app.TreasuryKeeper.GetAllBudgets(paramsContext())}, nil
	})
	common.RegisterQueryRouteWithParam(mux, "GET", "/nexarail/treasury/v1/grant/{id}", "id", func(id string) (interface{}, error) {
		if id == "" {
			return nil, fmt.Errorf("grant id required")
		}
		grant, found := app.TreasuryKeeper.GetGrant(paramsContext(), id)
		if !found {
			return nil, fmt.Errorf("grant %s not found", id)
		}
		return map[string]interface{}{"grant": grant}, nil
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/treasury/v1/grants", func() (interface{}, error) {
		return map[string]interface{}{"grants": app.TreasuryKeeper.GetAllGrants(paramsContext())}, nil
	})
}

func (app *NexaRailApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

func (app *NexaRailApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

func (app *NexaRailApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

func (app *NexaRailApp) RegisterGRPCServer(srv gogogrpc.Server) {
	app.BaseApp.RegisterGRPCServer(srv)
}

func (app *NexaRailApp) CommitMultiStore() storetypes.CommitMultiStore {
	return app.BaseApp.CommitMultiStore()
}

func (app *NexaRailApp) SnapshotManager() *snapshots.Manager {
	return nil
}

func (app *NexaRailApp) Close() error {
	return nil
}

// Phase 8F: registerUpgradeHandlers registers no-op upgrade handlers for testnet use.
// These prove the upgrade infrastructure works without mutating state.
// WARNING: Do not register real state-mutating handlers without full audit.
func (app *NexaRailApp) registerUpgradeHandlers() {
	// v0.2.0-testnet: future upgrade placeholder — no state mutation
	app.UpgradeKeeper.SetUpgradeHandler("v0.2.0-testnet", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("No-op upgrade handler executed", "plan", plan.Name, "height", plan.Height)
		// No state mutation — return current version map unchanged
		return fromVM, nil
	})
}

func GetMaccPerms() map[string][]string {
	return map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		crisistypes.ModuleName:         nil,
		merchanttypes.ModuleName:       nil,
		NexaRailEscrowModuleAccount:    nil,
		NexaRailTreasuryModuleAccount:  nil,
		NexaRailFeeRouterModuleAccount: nil,
		NexaRailBurnerModuleAccount:    {authtypes.Burner},
	}
}

func initParamsKeeper(
	appCodec codec.BinaryCodec,
	legacyAmino *codec.LegacyAmino,
	key, tkey storetypes.StoreKey,
) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)
	paramsKeeper.Subspace(baseapp.Paramspace)
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	return paramsKeeper
}

func SetBech32Prefix() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(AccountAddressPrefix, AccountAddressPrefix+"pub")
	cfg.SetBech32PrefixForValidator(AccountAddressPrefix+"valoper", AccountAddressPrefix+"valoperpub")
	cfg.SetBech32PrefixForConsensusNode(AccountAddressPrefix+"valcons", AccountAddressPrefix+"valconspub")
	cfg.SetCoinType(118)
	cfg.SetPurpose(44)
}

// paramStore is a minimal in-memory ParamStore for consensus parameters.
//
// BaseApp expects Get to return non-nil consensus params during proposal
// preparation and processing. InitChain calls Set on fresh chains, but after a
// process restart this in-memory store is rebuilt. Seed it from genesis/defaults
// so restart paths do not expose nil params to BaseApp.
type paramStore struct {
	cp *tmproto.ConsensusParams
}

func newParamStore(homePath string) *paramStore {
	cp := defaultConsensusParams(homePath)
	return &paramStore{cp: cloneConsensusParams(&cp)}
}

func defaultConsensusParams(homePath string) tmproto.ConsensusParams {
	if homePath != "" {
		genFile := filepath.Join(homePath, "config", "genesis.json")
		if genDoc, err := cmttypes.GenesisDocFromFile(genFile); err == nil && genDoc.ConsensusParams != nil {
			return genDoc.ConsensusParams.ToProto()
		}
	}

	return cmttypes.DefaultConsensusParams().ToProto()
}

func (ps *paramStore) Get(_ sdk.Context) (*tmproto.ConsensusParams, error) {
	if ps.cp == nil {
		cp := defaultConsensusParams("")
		ps.cp = cloneConsensusParams(&cp)
	}
	return cloneConsensusParams(ps.cp), nil
}

func (ps *paramStore) Has(_ sdk.Context) bool {
	return true
}

func (ps *paramStore) Set(_ sdk.Context, cp *tmproto.ConsensusParams) {
	if cp == nil {
		defaults := defaultConsensusParams("")
		ps.cp = cloneConsensusParams(&defaults)
		return
	}
	ps.cp = cloneConsensusParams(cp)
}

func cloneConsensusParams(cp *tmproto.ConsensusParams) *tmproto.ConsensusParams {
	if cp == nil {
		defaults := defaultConsensusParams("")
		cp = &defaults
	}

	out := &tmproto.ConsensusParams{}
	if cp.Block != nil {
		block := *cp.Block
		out.Block = &block
	}
	if cp.Evidence != nil {
		evidence := *cp.Evidence
		out.Evidence = &evidence
	}
	if cp.Validator != nil {
		validator := *cp.Validator
		validator.PubKeyTypes = append([]string(nil), cp.Validator.PubKeyTypes...)
		out.Validator = &validator
	}
	if cp.Version != nil {
		version := *cp.Version
		out.Version = &version
	}
	return out
}
