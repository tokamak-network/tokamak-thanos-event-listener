package main

import (
	"os"

	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"

	"github.com/urfave/cli/v2"

	thanosnotif "github.com/tokamak-network/tokamak-thanos-event-listener/internal/app/thanos-notif"
)

var app = &cli.App{
	Name:  "thanos-notif",
	Usage: "The thanos-notif command line interface",
}

func init() {
	app.Action = setupThanosListener
	app.Flags = append(app.Flags, thanosnotif.Flags()...)
}

func main() {
	log.GetLogger().Info("Start the application")
	if err := app.Run(os.Args); err != nil {
		log.GetLogger().Fatalw("Failed to start the application", "err", err)
	}
}

func setupThanosListener(ctx *cli.Context) error {
	config := &thanosnotif.Config{
		L1Rpc:                  ctx.String(thanosnotif.L1RpcUrlFlagName),
		L1WsRpc:                ctx.String(thanosnotif.L1WsRpcUrlFlagName),
		L2Rpc:                  ctx.String(thanosnotif.L2RpcUrFlagName),
		L2WsRpc:                ctx.String(thanosnotif.L2WsRpcUrFlagName),
		L1StandBridge:          ctx.String(thanosnotif.L1StandardBridgeFlagName),
		L2StandBridge:          ctx.String(thanosnotif.L2StandardBridgeFlagName),
		L1CrossDomainMessenger: ctx.String(thanosnotif.L1CrossDomainMessengerFlagName),
		L2CrossDomainMessenger: ctx.String(thanosnotif.L2CrossDomainMessengerFlagName),
		L2ToL1MessagePasser:    ctx.String(thanosnotif.L2ToL1MessengerParserFlagName),
		OptimismPortal:         ctx.String(thanosnotif.OptimismPortalFlagName),
		SlackURL:               ctx.String(thanosnotif.SlackUrlFlagName),
		TransferEventAddresses: ctx.StringSlice(thanosnotif.TransferAddressesFlagName),
	}

	log.GetLogger().Infow("Set up configuration", "config", config)

	app := thanosnotif.New(config)
	return app.Start()
}
