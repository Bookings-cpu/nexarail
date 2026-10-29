package cmd

import (
	"io"
	"os"
	"path/filepath"

	dbm "github.com/cometbft/cometbft-db"
	tmconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	paramscli "github.com/cosmos/cosmos-sdk/x/params/client/cli"
	slashingcli "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"

	"github.com/nexarail/chain/app"
	escrowcli "github.com/nexarail/chain/x/escrow/client/cli"
	feescli "github.com/nexarail/chain/x/fees/client/cli"
	merchantcli "github.com/nexarail/chain/x/merchant/client/cli"
	payoutcli "github.com/nexarail/chain/x/payout/client/cli"
	settlementcli "github.com/nexarail/chain/x/settlement/client/cli"
	treasurycli "github.com/nexarail/chain/x/treasury/client/cli"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "nexaraild",
		Short: "NexaRail blockchain daemon",
		Long: `NexaRail is a sovereign Layer 1 blockchain for railway settlement and payments.

Built on Cosmos SDK, NexaRail provides a controlled testnet-stage payment protocol for the rail industry
with external validator distribution still pending
with capabilities for merchant settlement, fee splitting, escrow, and treasury management.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			encodingConfig := app.MakeEncodingConfig()

			// Base client context. Home is resolved from --home (or default).
			clientCtx := sdkclient.Context{}.
				WithCodec(encodingConfig.Codec).
				WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
				WithTxConfig(encodingConfig.TxConfig).
				WithLegacyAmino(encodingConfig.Amino).
				WithInput(os.Stdin).
				WithAccountRetriever(types.AccountRetriever{}).
				WithHomeDir(app.DefaultNodeHome).
				WithViper("")

			// Apply persistent CLI flags (--home, --chain-id, --node, etc.),
			// then merge in values from client.toml. ReadFromClientConfig
			// creates the config dir and a default client.toml when missing,
			// so bootstrap commands (init, keys, gentx) work from a fresh home.
			clientCtx, err := sdkclient.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientCtx, err = clientconfig.ReadFromClientConfig(clientCtx)
			if err != nil {
				return err
			}

			if err := sdkclient.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			// Load config.toml and app.toml from disk into the server context.
			// This is what applies p2p.persistent_peers, p2p.laddr, rpc.laddr,
			// addr_book_strict and allow_duplicate_ip to the running node.
			// InterceptConfigsPreRunHandler writes default files when they don't
			// exist, so bootstrap commands are safe here too.
			return server.InterceptConfigsPreRunHandler(cmd, "", nil, tmconfig.DefaultConfig())
		},
	}

	rootCmd.PersistentFlags().String(flags.FlagLogLevel, "info", "log level")
	rootCmd.PersistentFlags().String(flags.FlagLogFormat, "plain", "log format")
	rootCmd.PersistentFlags().String(flags.FlagChainID, "nexarail-devnet-1", "chain ID")
	rootCmd.PersistentFlags().String(flags.FlagHome, app.DefaultNodeHome, "directory for config and data")

	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, genutiltypes.DefaultMessageValidator),
		genutilcli.GenTxCmd(app.ModuleBasics, app.MakeEncodingConfig().TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		keys.Commands(app.DefaultNodeHome),
		server.StartCmd(newAppCreator(), app.DefaultNodeHome),
		server.ExportCmd(exportAppStateAndValidators, app.DefaultNodeHome),
		server.VersionCmd(),
		tendermintCommand(),
		DebugP2PConfigCmd(),
		DebugLiveFlagsCmd(),
		DebugModuleSummaryCmd(),
		rpc.StatusCommand(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		queryCommand(),
		txCommand(),
	)

	return rootCmd
}

func newAppCreator() servertypes.AppCreator {
	return func(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
		homePath, ok := appOpts.Get(flags.FlagHome).(string)
		if !ok || homePath == "" {
			homePath = app.DefaultNodeHome
		}

		encodingConfig := app.MakeEncodingConfig()
		app.RegisterInterfaces(encodingConfig.InterfaceRegistry)

		// Read chain ID from genesis.json so BaseApp.InitChain validates correctly.
		genFile := filepath.Join(homePath, "config", "genesis.json")
		chainID := ""
		if genDoc, err := tmtypes.GenesisDocFromFile(genFile); err == nil {
			chainID = genDoc.ChainID
		}

		baseAppOpts := server.DefaultBaseappOptions(appOpts)
		baseAppOpts = append(baseAppOpts, baseapp.SetChainID(chainID))

		return app.NewNexaRailApp(
			logger,
			db,
			traceStore,
			true,
			nil,
			homePath,
			0,
			encodingConfig,
			appOpts,
			baseAppOpts...,
		)
	}
}

func exportAppStateAndValidators(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		homePath = app.DefaultNodeHome
	}

	encodingConfig := app.MakeEncodingConfig()
	app.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	nexApp := app.NewNexaRailApp(
		logger,
		db,
		traceStore,
		height == -1,
		nil,
		homePath,
		0,
		encodingConfig,
		appOpts,
	)

	if height != -1 {
		if err := nexApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return nexApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}
	cmd.PersistentFlags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
	cmd.PersistentFlags().String(flags.FlagGRPC, "", "the gRPC endpoint to use for this chain")
	cmd.PersistentFlags().Bool(flags.FlagGRPCInsecure, false, "allow gRPC over insecure channels, if not TLS the server must use TLS")
	cmd.PersistentFlags().Int64(flags.FlagHeight, 0, "Use a specific height to query state at")
	cmd.PersistentFlags().StringP(flags.FlagOutput, "o", "text", "Output format (text|json)")

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcli.GetQueryCmd(),
		bankcli.GetQueryCmd(),
		govcli.GetQueryCmd(),
		stakingcli.GetQueryCmd(),
		slashingcli.GetQueryCmd(),
		feescli.GetQueryCmd(),
		merchantcli.GetQueryCmd(),
		settlementcli.GetQueryCmd(),
		escrowcli.GetQueryCmd(),
		treasurycli.GetQueryCmd(),
		payoutcli.GetQueryCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		bankcli.NewSendTxCmd(),
		govcli.NewTxCmd(nil),
		stakingcli.NewTxCmd(),
		slashingcli.NewTxCmd(),
		paramscli.NewSubmitParamChangeProposalTxCmd(),
		feescli.GetTxCmd(),
		merchantcli.GetTxCmd(),
		settlementcli.GetTxCmd(),
		escrowcli.GetTxCmd(),
		treasurycli.GetTxCmd(),
		payoutcli.GetTxCmd(),
		authcli.GetSignCommand(),
		authcli.GetBroadcastCommand(),
		authcli.GetEncodeCommand(),
		authcli.GetDecodeCommand(),
	)

	return cmd
}
