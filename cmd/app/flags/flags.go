package flags

import (
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/urfave/cli/v2"
)

const (
	NetworkFlagName          = "network"
	L1RpcUrlFlagName         = "l1-rpc"
	L1WsRpcUrlFlagName       = "l1-ws-rpc"
	L2RpcUrlFlagName         = "l2-rpc"
	L2WsRpcUrlFlagName       = "l2-ws-rpc"
	L1StandardBridgeFlagName = "l1-standard-bridge-address"
	L2StandardBridgeFlagName = "l2-standard-bridge-address"
	SlackUrlFlagName         = "slack-url"
	L1ExplorerUrlFlagName    = "l1-explorer-url"
	L2ExplorerUrlFlagName    = "l2-explorer-url"
	OffFlagName              = "slack-on-off"
	TokenAddressesFlagName   = "token-addresses"
)

var (
	NetworkFlag = &cli.StringFlag{
		Name:    NetworkFlagName,
		Usage:   "Network name",
		EnvVars: []string{"NETWORK"},
	}
	L1RpcFlag = &cli.StringFlag{
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
	L2WsRpcFlag = &cli.StringFlag{
		Name:    L2WsRpcUrlFlagName,
		Usage:   "L2 Ws RPC url",
		Value:   "ws://localhost:9546",
		EnvVars: []string{"L2_WS_RPC"},
	}
	L1StandardBridgeFlag = &cli.StringFlag{
		Name:    L1StandardBridgeFlagName,
		Usage:   "L1StandardBridge address",
		EnvVars: []string{"L1_STANDARD_BRIDGE"},
	}
	L2StandardBridgeFlag = &cli.StringFlag{
		Name:    L2StandardBridgeFlagName,
		Usage:   "L2StandardBridge address",
		Value:   predeploys.L2StandardBridge,
		EnvVars: []string{"L2_STANDARD_BRIDGE"},
	}
	SlackUrlFlag = &cli.StringFlag{
		Name:    SlackUrlFlagName,
		Usage:   "slack url for notification",
		EnvVars: []string{"SLACK_URL"},
	}
	L1ExplorerUrlFlag = &cli.StringFlag{
		Name:    L1ExplorerUrlFlagName,
		Usage:   "L1 explorer url",
		EnvVars: []string{"L1_EXPLORER_URL"},
	}
	L2ExplorerUrlFlag = &cli.StringFlag{
		Name:    L2ExplorerUrlFlagName,
		Usage:   "L2 explorer url",
		EnvVars: []string{"L2_EXPLORER_URL"},
	}
	OffFlag = &cli.BoolFlag{
		Name:    OffFlagName,
		Usage:   "Slack active",
		EnvVars: []string{"OFF"},
	}
	TokenAddressesFlag = &cli.StringSliceFlag{
		Name:    TokenAddressesFlagName,
		Usage:   "List of addresses to get symbol and decimals",
		EnvVars: []string{"TOKEN_ADDRESSES"},
	}
)

func Flags() []cli.Flag {
	return []cli.Flag{
		NetworkFlag,
		L1RpcFlag,
		L1WsRpcFlag,
		L2RPCFlag,
		L2WsRpcFlag,
		L1StandardBridgeFlag,
		L2StandardBridgeFlag,
		SlackUrlFlag,
		L1ExplorerUrlFlag,
		L2ExplorerUrlFlag,
		OffFlag,
		TokenAddressesFlag,
	}
}
