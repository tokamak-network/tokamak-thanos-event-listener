package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tokamak-network/tokamak-thanos-event-listener/cmd/utils"
	notif "github.com/tokamak-network/tokamak-thanos-event-listener/thanos-notif"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "thanos-notif",
	Usage: "The thanos-notif command line interface",
}

func init() {
	app.Action = setupThanosListener
	app.Flags = append(app.Flags, []cli.Flag{
		utils.L1RPCFlag,
		utils.L2RPCFlag,
		utils.L1StandardBridgeFlag,
		utils.L2StandardBridgeFlag,
		utils.L1CrossDomainMessengerFlag,
		utils.L2CrossDomainMessengerFlag,
		utils.L2ToL1MessagePasserFlag,
		utils.OptimismPortalFlag,
		utils.SlackUrlFlag,
	}...)
	fmt.Println("Init:")
}

func main() {
	fmt.Println("thanos-notif")
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setupThanosListener(ctx *cli.Context) error {
	config := &notif.NotifAppConfig{
		L1_RPC:                 ctx.String("l1-rpc"),
		L2_RPC:                 ctx.String("l2-rpc"),
		L1StandBridge:          ctx.String("l1-standard-bridge-address"),
		L2StandBridge:          ctx.String("l2-standard-bridge-address"),
		L1CrossDomainMessenger: ctx.String("l1-cross-domain-messenger-address"),
		L2CrossDomainMessenger: ctx.String("l2-cross-domain-messenger-address"),
		L2ToL1MessagePasser:    ctx.String("l2-to-l1-message-parser-address"),
		OptimismPortal:         ctx.String("optimism-portal-address"),
		SlackURL:               ctx.String("slack-url"),
	}
	fmt.Println("[config]:", config)
	app := notif.MakeNotifApp(config)
	return app.Start()
}
