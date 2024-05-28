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
	app.Flags = append(app.Flags, Flags()...)
}

func main() {
	log.GetLogger().Info("Start the application")
	if err := app.Run(os.Args); err != nil {
		log.GetLogger().Fatalw("Failed to start the application", "err", err)
	}
}

func setupThanosListener(ctx *cli.Context) error {
	config := &thanosnotif.Config{
		L1Rpc:                  ctx.String(L1RpcUrlFlagName),
		L1WsRpc:                ctx.String(L1WsRpcUrlFlagName),
		L2Rpc:                  ctx.String(L2RpcUrFlagName),
		L2WsRpc:                ctx.String(L2WsRpcUrFlagName),
		L1StandBridge:          ctx.String(L1StandardBridgeFlagName),
		L2StandBridge:          ctx.String(L2StandardBridgeFlagName),
		L1CrossDomainMessenger: ctx.String(L1CrossDomainMessengerFlagName),
		L2CrossDomainMessenger: ctx.String(L2CrossDomainMessengerFlagName),
		L2ToL1MessagePasser:    ctx.String(L2ToL1MessengerParserFlagName),
		OptimismPortal:         ctx.String(OptimismPortalFlagName),
		SlackURL:               ctx.String(SlackUrlFlagName),
		TransferEventAddresses: ctx.StringSlice(TransferAddressesFlagName),
	}

	log.GetLogger().Infow("Set up configuration", "config", config)

	app := thanosnotif.New(config)
	return app.Start()
}