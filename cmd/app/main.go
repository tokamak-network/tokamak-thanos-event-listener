package main

import (
	"os"

	"github.com/tokamak-network/tokamak-thanos-event-listener/cmd/app/flags"
	thanosnotif "github.com/tokamak-network/tokamak-thanos-event-listener/internal/app/thanos-notif"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "thanos-notif",
		Usage: "The thanos-notif command line interface",
		Flags: flags.Flags(),
		Commands: []*cli.Command{
			{
				Name:    "listener",
				Aliases: []string{},
				Action:  startListener,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.GetLogger().Fatalw("Failed to start the application", "err", err)
	}
}

func startListener(ctx *cli.Context) error {
	log.GetLogger().Info("Start the application")

	config := &thanosnotif.Config{
		L1Rpc:                  ctx.String(flags.L1RpcUrlFlagName),
		L1WsRpc:                ctx.String(flags.L1WsRpcUrlFlagName),
		L2Rpc:                  ctx.String(flags.L2RpcUrlFlagName),
		L2WsRpc:                ctx.String(flags.L2WsRpcUrlFlagName),
		L1StandardBridge:       ctx.String(flags.L1StandardBridgeFlagName),
		L2StandardBridge:       ctx.String(flags.L2StandardBridgeFlagName),
		L1CrossDomainMessenger: ctx.String(flags.L1CrossDomainMessengerFlagName),
		L2CrossDomainMessenger: ctx.String(flags.L2CrossDomainMessengerFlagName),
		L2ToL1MessagePasser:    ctx.String(flags.L2ToL1MessengerPasserFlagName),
		OptimismPortal:         ctx.String(flags.OptimismPortalFlagName),
		SlackURL:               ctx.String(flags.SlackUrlFlagName),
		TransferEventAddresses: ctx.StringSlice(flags.TransferAddressesFlagName),
		OFF:                    ctx.Bool(flags.OffFlagName),
	}

	log.GetLogger().Infow("Set up configuration", "config", config)

	app := thanosnotif.New(config)

	return app.Start()
}
