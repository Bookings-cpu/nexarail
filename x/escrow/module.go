package escrow

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
	"github.com/nexarail/chain/x/escrow/client/cli"
	"github.com/nexarail/chain/x/escrow/keeper"
	"github.com/nexarail/chain/x/escrow/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}
func (AppModuleBasic) RegisterInterfaces(r cdctypes.InterfaceRegistry) { types.RegisterInterfaces(r) }

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	bz, _ := json.Marshal(types.DefaultGenesis())
	return bz
}
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return fmt.Errorf("escrow genesis: %w", err)
	}
	return gs.Validate()
}
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router)              {}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	common.RegisterQueryRoute(mux, "GET", "/nexarail/escrow/v1/params", func() (interface{}, error) {
		qc := types.NewQueryClient(clientCtx)
		return qc.Params(context.Background(), &types.QueryParamsRequest{})
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/escrow/v1/escrows", func() (interface{}, error) {
		qc := types.NewQueryClient(clientCtx)
		return qc.Escrows(context.Background(), &types.QueryEscrowsRequest{})
	})
}
func (AppModuleBasic) GetTxCmd() *cobra.Command                                        { return cli.GetTxCmd() }
func (AppModuleBasic) GetQueryCmd() *cobra.Command                                     { return cli.GetQueryCmd() }

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

// RegisterInvariants registers the escrow module invariants. This was
// previously a no-op (`func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}`),
// so ValidateCustodyInvariant existed and was correct but never actually ran —
// the crisis module's periodic/manual invariant checks had nothing to check.
func RegisterInvariants(ir sdk.InvariantRegistry, k keeper.Keeper) {
	ir.RegisterRoute(types.ModuleName, "custody", CustodyInvariant(k))
}

// CustodyInvariant checks that no terminal escrow still shows funds in custody.
func CustodyInvariant(k keeper.Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		err := k.ValidateCustodyInvariant(ctx)
		broken := err != nil
		msg := "escrow custody invariant OK"
		if broken {
			msg = err.Error()
		}
		return sdk.FormatInvariant(types.ModuleName, "custody", msg), broken
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	keeper.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	keeper.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var gs types.GenesisState
	if err := json.Unmarshal(data, &gs); err != nil {
		panic(fmt.Errorf("escrow genesis: %w", err))
	}
	// Previously only validated by the standalone `validate-genesis` CLI path
	// (AppModuleBasic.ValidateGenesis), never by InitGenesis itself — so a
	// hand-edited or migration-produced genesis could load unvalidated state,
	// including an escrow with a malformed address that only panicked later
	// the first time a release/refund/dispute tried to pay it out.
	if err := gs.Validate(); err != nil {
		panic(fmt.Errorf("escrow genesis invalid: %w", err))
	}
	if err := am.keeper.SetParams(ctx, gs.Params); err != nil {
		panic(fmt.Errorf("escrow params: %w", err))
	}
	for _, e := range gs.Escrows {
		if err := am.keeper.SetEscrow(ctx, e); err != nil {
			panic(fmt.Errorf("escrow %s: %w", e.EscrowId, err))
		}
	}
	am.keeper.RebuildIndexes(ctx)
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	bz, _ := json.Marshal(&types.GenesisState{Params: am.keeper.GetParams(ctx), Escrows: am.keeper.GetAllEscrows(ctx)})
	return bz
}
func (AppModule) ConsensusVersion() uint64                                  { return 1 }
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {}
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return nil
}
