package flags

import (
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/urfave/cli/v2"
)

const (
	NetworkFlagName          = "network"
	L1HttpRpcUrlFlagName     = "l1-http-rpc-url"
	L1WsRpcUrlFlagName       = "l1-ws-rpc"
	L2WsRpcUrlFlagName       = "l2-ws-rpc"
	L2HttpRpcUrlFlagName     = "l2-http-rpc"
	OptimismPortalFlagName   = "optimism-portal"
	L1XDMFlagName            = "l1-xdm-address"
	L1StandardBridgeFlagName = "l1-standard-bridge-address"
	L2StandardBridgeFlagName = "l2-standard-bridge-address"
	L1UsdcBridgeFlagName     = "l1-usdc-bridge-address"
	L2UsdcBridgeFlagName     = "l2-usdc-bridge-address"
	SlackUrlFlagName         = "slack-url"
	L1ExplorerUrlFlagName    = "l1-explorer-url"
	L2ExplorerUrlFlagName    = "l2-explorer-url"
	L1TokenAddresses         = "l1-token-addresses"
	L2TokenAddresses         = "l2-token-addresses"
	RedisAddressFlagName     = "redis-address"
	RedisDBFlagName          = "redis-db"
)

var (
	NetworkFlag = &cli.StringFlag{
		Name:    NetworkFlagName,
		Usage:   "Network name",
		EnvVars: []string{"NETWORK"},
	}
	L1HttpRpcFlag = &cli.StringFlag{
		Name:    L1HttpRpcUrlFlagName,
		Usage:   "L1 HTTP RPC url",
		Value:   "http://localhost:8545",
		EnvVars: []string{"L1_HTTP_RPC"},
	}
	L1WsRpcFlag = &cli.StringFlag{
		Name:    L1WsRpcUrlFlagName,
		Usage:   "L1 RPC url",
		Value:   "ws://localhost:8546",
		EnvVars: []string{"L1_WS_RPC"},
	}
	L2WsRpcFlag = &cli.StringFlag{
		Name:    L2WsRpcUrlFlagName,
		Usage:   "L2 Ws RPC url",
		Value:   "ws://localhost:9546",
		EnvVars: []string{"L2_WS_RPC"},
	}
	L2HttpRpcFlag = &cli.StringFlag{
		Name:    L2HttpRpcUrlFlagName,
		Usage:   "L2 HTTP RPC url",
		Value:   "http://localhost:9545",
		EnvVars: []string{"L2_HTTP_RPC"},
	}
	OptimismPortalFlag = &cli.StringFlag{
		Name:    OptimismPortalFlagName,
		Usage:   "OptimismPortal address",
		EnvVars: []string{"OPTIMISM_PORTAL"},
	}
	L1XDMFlag = &cli.StringFlag{
		Name:    L1XDMFlagName,
		Usage:   "L1CrossDomainMessenger address",
		EnvVars: []string{"L1_XDM"},
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
	L1UsdcBridgeFlag = &cli.StringFlag{
		Name:    L1UsdcBridgeFlagName,
		Usage:   "L1UsdcBridge address",
		EnvVars: []string{"L1_USDC_BRIDGE"},
	}
	L2UsdcBridgeFlag = &cli.StringFlag{
		Name:    L2UsdcBridgeFlagName,
		Usage:   "L2UsdcBridge address",
		EnvVars: []string{"L2_USDC_BRIDGE"},
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

	L1TokenAddressesFlag = &cli.StringSliceFlag{
		Name:    L1TokenAddresses,
		Usage:   "List of L1 tokens address to get symbol and decimals",
		EnvVars: []string{"L1_TOKEN_ADDRESSES"},
	}
	L2TokenAddressesFlag = &cli.StringSliceFlag{
		Name:    L2TokenAddresses,
		Usage:   "List of L2 tokens address to get symbol and decimals",
		EnvVars: []string{"L2_TOKEN_ADDRESSES"},
	}
	RedisAddressFlag = &cli.StringFlag{
		Name: RedisAddressFlagName,
		EnvVars: []string{
			"REDIS_ADDRESS",
		},
	}
	RedisDBFlag = &cli.IntFlag{
		Name: RedisDBFlagName,
		EnvVars: []string{
			"REDIS_DB",
		},
	}
)

func Flags() []cli.Flag {
	return []cli.Flag{
		NetworkFlag,
		L1WsRpcFlag,
		L1HttpRpcFlag,
		L2WsRpcFlag,
		L2HttpRpcFlag,
		OptimismPortalFlag,
		L1XDMFlag,
		L1StandardBridgeFlag,
		L2StandardBridgeFlag,
		L1UsdcBridgeFlag,
		L2UsdcBridgeFlag,
		SlackUrlFlag,
		L1ExplorerUrlFlag,
		L2ExplorerUrlFlag,
		L1TokenAddressesFlag,
		L2TokenAddressesFlag,
		RedisAddressFlag,
		RedisDBFlag,
	}
}
