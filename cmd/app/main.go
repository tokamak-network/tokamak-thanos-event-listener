package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/tokamak-network/tokamak-thanos-event-listener/cmd/app/flags"
	thanosnotif "github.com/tokamak-network/tokamak-thanos-event-listener/internal/app/thanos-notif"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/redis"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
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
		log.GetLogger().Fatalw("Failed to start the application", "error", err)
	}
}

func startListener(ctx *cli.Context) error {
	log.GetLogger().Info("Start the application")

	config := &thanosnotif.Config{
		Network:          ctx.String(flags.NetworkFlagName),
		L1WsRpc:          ctx.String(flags.L1WsRpcUrlFlagName),
		L2WsRpc:          ctx.String(flags.L2WsRpcUrlFlagName),
		L1StandardBridge: ctx.String(flags.L1StandardBridgeFlagName),
		L2StandardBridge: ctx.String(flags.L2StandardBridgeFlagName),
		L1UsdcBridge:     ctx.String(flags.L1UsdcBridgeFlagName),
		L2UsdcBridge:     ctx.String(flags.L2UsdcBridgeFlagName),
		SlackURL:         ctx.String(flags.SlackUrlFlagName),
		L1ExplorerUrl:    ctx.String(flags.L1ExplorerUrlFlagName),
		L2ExplorerUrl:    ctx.String(flags.L2ExplorerUrlFlagName),
		L1TokenAddresses: ctx.StringSlice(flags.L1TokenAddresses),
		L2TokenAddresses: ctx.StringSlice(flags.L2TokenAddresses),
		RedisConfig: redis.Config{
			Addresses: ctx.String(flags.RedisAddressFlagName),
		},
	}

	if err := config.Validate(); err != nil {
		log.GetLogger().Fatalw("Failed to start the application", "error", err)
	}

	log.GetLogger().Infow("Set up configuration", "config", config)

	app, err := thanosnotif.New(ctx.Context, config)
	if err != nil {
		log.GetLogger().Errorw("Failed to start the application", "error", err)
		return err
	}

	return app.Start(ctx.Context)
}
