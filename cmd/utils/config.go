package utils

import (
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/urfave/cli/v2"
)

var (
	L1RPCFlag = &cli.StringFlag{
		Name:    "l1-rpc",
		Usage:   "L1 RPC url",
		Value:   "http://localhost:8545",
		EnvVars: []string{"L1_RPC"},
	}
	L2RPCFlag = &cli.StringFlag{
		Name:    "l2-rpc",
		Usage:   "L2 RPC url",
		Value:   "http://localhost:9545",
		EnvVars: []string{"L2_RPC"},
	}
	L1StandardBridgeFlag = &cli.StringFlag{
		Name:    "l1-standard-bridge-address",
		Usage:   "L1StandBridge address",
		EnvVars: []string{"L1_STANDARD_BIRDGE"},
	}
	L2StandardBridgeFlag = &cli.StringFlag{
		Name:    "l2-standard-bridge-address",
		Usage:   "L2StandBridge address",
		Value:   predeploys.L2StandardBridge,
		EnvVars: []string{"L2_STANDARD_BIRDGE"},
	}
	L1CrossDomainMessengerFlag = &cli.StringFlag{
		Name:    "l1-cross-domain-messenger-address",
		Usage:   "L1CrossDomainMessenger address",
		EnvVars: []string{"L1_CROSS_DOMAIN_MESSENGER"},
	}
	L2CrossDomainMessengerFlag = &cli.StringFlag{
		Name:    "l2-cross-domain-messenger-address",
		Usage:   "L2CrossDomainMessenger address",
		Value:   predeploys.L2CrossDomainMessenger,
		EnvVars: []string{"L2_CROSS_DOMAIN_MESSENGER"},
	}
	L2ToL1MessagePasserFlag = &cli.StringFlag{
		Name:    "l2-to-l1-message-parser-address",
		Usage:   "L2ToL1MessagePasser address",
		Value:   predeploys.L2ToL1MessagePasser,
		EnvVars: []string{"L2_TO_L1_MESSAGE_PASSER"},
	}
	OptimismPortalFlag = &cli.StringFlag{
		Name:    "optimism-portal-address",
		Usage:   "OptimismPortal address",
		EnvVars: []string{"OPTIMISM_PORTAL"},
	}
	SlackUrlFlag = &cli.StringFlag{
		Name:    "slack-url",
		Usage:   "slack url for notification",
		EnvVars: []string{"OPTIMISM_PORTAL"},
	}
)
