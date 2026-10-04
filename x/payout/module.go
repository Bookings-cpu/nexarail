package payout

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
	"github.com/nexarail/chain/x/payout/client/cli"
	"github.com/nexarail/chain/x/payout/keeper"
	"github.com/nexarail/chain/x/payout/types"
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
		return fmt.Errorf("payout genesis: %w", err)
	}
	return gs.Validate()
}
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router)              {}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	common.RegisterQueryRoute(mux, "GET", "/nexarail/payout/v1/params", func() (interface{}, error) {
		qc := types.NewQueryClient(clientCtx)
		return qc.Params(context.Background(), &types.QueryParamsRequest{})
	})
	common.RegisterQueryRoute(mux, "GET", "/nexarail/payout/v1/payouts", func() (interface{}, error) {
		qc := types.NewQueryClient(clientCtx)
		return qc.Payouts(context.Background(), &types.QueryPayoutsRequest{})
	})
}
func (AppModuleBasic) GetTxCmd() *cobra.Command                                        { return cli.GetTxCmd() }
func (AppModuleBasic) GetQueryCmd() *cobra.Command                                     { return cli.GetQueryCmd() }

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(k keeper.Keeper) AppModule                    { return AppModule{AppModuleBasic{}, k} }
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	RegisterInvariants(ir, am.keeper)
}

// RegisterInvariants registers the payout module invariants. Previously a
// no-op — ValidatePayoutFundsInvariant existed and was correct but never ran.
func RegisterInvariants(ir sdk.InvariantRegistry, k keeper.Keeper) {
	ir.RegisterRoute(types.ModuleName, "funds", FundsInvariant(k))
}

// FundsInvariant checks FundsPaid/status consistency across all payouts.
func FundsInvariant(k keeper.Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		err := k.ValidatePayoutFundsInvariant(ctx)
		broken := err != nil
		msg := "payout funds invariant OK"
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
		panic(err)
	}
	// Previously only validated by the standalone `validate-genesis` CLI path,
	// never by InitGenesis itself — a hand-edited or migration-produced
	// genesis could load unvalidated payout/batch records.
	if err := gs.Validate(); err != nil {
		panic(fmt.Errorf("payout genesis invalid: %w", err))
	}
	if err := am.keeper.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}
	for _, p := range gs.Payouts {
		am.keeper.SetPayout(ctx, p)
	}
	for _, b := range gs.BatchPayouts {
		am.keeper.SetBatchPayout(ctx, b)
	}
	am.keeper.RebuildIndexes(ctx)
	return nil
}
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	bz, _ := json.Marshal(&types.GenesisState{Params: am.keeper.GetParams(ctx), Payouts: am.keeper.GetAllPayouts(ctx), BatchPayouts: am.keeper.GetAllBatchPayouts(ctx)})
	return bz
}
func (AppModule) ConsensusVersion() uint64                                  { return 1 }
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {}
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return nil
}
