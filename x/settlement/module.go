package settlement

import (
	"context"
	"encoding/json"
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"github.com/nexarail/chain/x/common"
	"github.com/nexarail/chain/x/settlement/client/cli"
	"github.com/nexarail/chain/x/settlement/keeper"
	"github.com/nexarail/chain/x/settlement/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// --- AppModuleBasic ---

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(types.DefaultGenesis())
	if err != nil {
		panic(err)
	}
	return bz
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := json.Unmarshal(bz, &data); err != nil {
		return fmt.Errorf("settlement genesis unmarshal: %w", err)
	}
	return data.Validate()
}

func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router)              {}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	common.RegisterQueryRoute(mux, "GET", "/nexarail/settlement/v1/params", func() (interface{}, error) {
		qc := types.NewQueryClient(clientCtx)
		return qc.Params(context.Background(), &types.QueryParamsRequest{})
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/settlement/v1/settlements", func() (interface{}, error) {
		qc := types.NewQueryClient(clientCtx)
		return qc.Settlements(context.Background(), &types.QuerySettlementsRequest{})
	})
}

func (AppModuleBasic) GetTxCmd() *cobra.Command    { return cli.GetTxCmd() }
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return cli.GetQueryCmd() }

// --- AppModule ---

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{AppModuleBasic: AppModuleBasic{}, keeper: k}
}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	RegisterInvariants(ir, am.keeper)
}

// RegisterInvariants registers the settlement module invariants. Previously
// a no-op — ValidateSettlementFundsInvariant existed and was correct but
// never ran.
func RegisterInvariants(ir sdk.InvariantRegistry, k keeper.Keeper) {
	ir.RegisterRoute(types.ModuleName, "funds", FundsInvariant(k))
}

// FundsInvariant checks fee + merchant-net consistency across all settlements.
func FundsInvariant(k keeper.Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		err := k.ValidateSettlementFundsInvariant(ctx)
		broken := err != nil
		msg := "settlement funds invariant OK"
		if broken {
			msg = err.Error()
		}
		return sdk.FormatInvariant(types.ModuleName, "funds", msg), broken
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	keeper.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	keeper.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var gs types.GenesisState
	if err := json.Unmarshal(data, &gs); err != nil {
		panic(fmt.Errorf("settlement genesis unmarshal: %w", err))
	}
	// Previously only validated by the standalone `validate-genesis` CLI path,
	// never by InitGenesis itself.
	if err := gs.Validate(); err != nil {
		panic(fmt.Errorf("settlement genesis invalid: %w", err))
	}
	if err := am.keeper.SetParams(ctx, gs.Params); err != nil {
		panic(fmt.Errorf("settlement genesis params: %w", err))
	}
	for _, s := range gs.Settlements {
		if err := am.keeper.SetSettlement(ctx, s); err != nil {
			panic(fmt.Errorf("settlement genesis record %d: %w", s.Id, err))
		}
	}
	// Rebuild merchant/payer indexes
	am.keeper.RebuildIndexes(ctx)
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	params := am.keeper.GetParams(ctx)
	settlements := am.keeper.GetAllSettlements(ctx)
	bz, err := json.Marshal(&types.GenesisState{Params: params, Settlements: settlements})
	if err != nil {
		panic(err)
	}
	return bz
}

func (AppModule) ConsensusVersion() uint64                                  { return 1 }
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {}
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
