package flags

import (
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/urfave/cli/v2"
)

const (
	NetworkFlagName                = "network"
	L1RpcUrlFlagName               = "l1-rpc"
	L1WsRpcUrlFlagName             = "l1-ws-rpc"
	L2RpcUrlFlagName               = "l2-rpc"
	L2WsRpcUrlFlagName             = "l2-ws-rpc"
	L1StandardBridgeFlagName       = "l1-standard-bridge-address"
	L2StandardBridgeFlagName       = "l2-standard-bridge-address"
	L1CrossDomainMessengerFlagName = "l1-cross-domain-messenger-address"
	L2CrossDomainMessengerFlagName = "l2-cross-domain-messenger-address"
	L2ToL1MessengerPasserFlagName  = "l2-to-l1-message-parser-address"
	OptimismPortalFlagName         = "optimism-portal-address"
	SlackUrlFlagName               = "slack-url"
	OffFlagName                    = "slack-on-off"
)

var (
	NetworkFlag = &cli.StringFlag{
		Name:    NetworkFlagName,
		Usage:   "Network name",
		EnvVars: []string{"NETWORK"},
	}
	L1RPCFlag = &cli.StringFlag{
		Name:    L1RpcUrlFlagName,
		Usage:   "L1 RPC url",
		Value:   "http://localhost:8545",
		EnvVars: []string{"L1_RPC"},
	}
	L1WsRpcFlag = &cli.StringFlag{
		Name:    L1WsRpcUrlFlagName,
		Usage:   "L1 RPC url",
		Value:   "ws://localhost:8546",
		EnvVars: []string{"L1_WS_RPC"},
	}
	L2RPCFlag = &cli.StringFlag{
		Name:    L2RpcUrlFlagName,
		Usage:   "L2 RPC url",
		Value:   "http://localhost:9545",
		EnvVars: []string{"L2_RPC"},
	}
	L2WsRPCFlag = &cli.StringFlag{
		Name:    L2WsRpcUrlFlagName,
		Usage:   "L2 Ws RPC url",
		Value:   "ws://localhost:9546",
		EnvVars: []string{"L2_WS_RPC"},
	}
	L1StandardBridgeFlag = &cli.StringFlag{
		Name:    L1StandardBridgeFlagName,
		Usage:   "L1StandardBridge address",
		EnvVars: []string{"L1_STANDARD_BIRDGE"},
	}
	L2StandardBridgeFlag = &cli.StringFlag{
		Name:    L2StandardBridgeFlagName,
		Usage:   "L2StandardBridge address",
		Value:   predeploys.L2StandardBridge,
		EnvVars: []string{"L2_STANDARD_BIRDGE"},
	}
	L1CrossDomainMessengerFlag = &cli.StringFlag{
		Name:    L1CrossDomainMessengerFlagName,
		Usage:   "L1CrossDomainMessenger address",
		EnvVars: []string{"L1_CROSS_DOMAIN_MESSENGER"},
	}
	L2CrossDomainMessengerFlag = &cli.StringFlag{
		Name:    L2CrossDomainMessengerFlagName,
		Usage:   "L2CrossDomainMessenger address",
		Value:   predeploys.L2CrossDomainMessenger,
		EnvVars: []string{"L2_CROSS_DOMAIN_MESSENGER"},
	}
	L2ToL1MessagePasserFlag = &cli.StringFlag{
		Name:    L2ToL1MessengerPasserFlagName,
		Usage:   "L2ToL1MessagePasser address",
		Value:   predeploys.L2ToL1MessagePasser,
		EnvVars: []string{"L2_TO_L1_MESSAGE_PASSER"},
	}
	OptimismPortalFlag = &cli.StringFlag{
		Name:    OptimismPortalFlagName,
		Usage:   "OptimismPortal address",
		EnvVars: []string{"OPTIMISM_PORTAL"},
	}
	SlackUrlFlag = &cli.StringFlag{
		Name:    SlackUrlFlagName,
		Usage:   "slack url for notification",
		EnvVars: []string{"SLACK_URL"},
	}
	OffFlag = &cli.BoolFlag{
		Name:    OffFlagName,
		Usage:   "Slack active",
		EnvVars: []string{"OFF"},
	}
)

func Flags() []cli.Flag {
	return []cli.Flag{
		NetworkFlag,
		L1RPCFlag,
		L1WsRpcFlag,
		L2RPCFlag,
		L2WsRPCFlag,
		L1StandardBridgeFlag,
		L2StandardBridgeFlag,
		L1CrossDomainMessengerFlag,
		L2CrossDomainMessengerFlag,
		L2ToL1MessagePasserFlag,
		OptimismPortalFlag,
		SlackUrlFlag,
		OffFlag,
	}
}
