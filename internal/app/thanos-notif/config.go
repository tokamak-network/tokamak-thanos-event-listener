package thanosnotif

import (
	"errors"
)

type Config struct {
	Network string

	L1Rpc   string
	L1WsRpc string

	L2Rpc   string
	L2WsRpc string

	L1StandardBridge string
	L2StandardBridge string

	L1UsdcBridge string
	L2UsdcBridge string

	SlackURL string

	L1ExplorerUrl string
	L2ExplorerUrl string

	OFF bool

	TokenAddresses []string

	TonAddress string
}

func (c *Config) Validate() error {
	if c.L1Rpc == "" {
		return errors.New("l1 rpc address is required")
	}

	if c.L1WsRpc == "" {
		return errors.New("l1 ws rpc address is required")
	}

	if c.L2Rpc == "" {
		return errors.New("l2 rpc address is required")
	}

	if c.L2WsRpc == "" {
		return errors.New("l2 ws rpc address is required")
	}

	if c.L1StandardBridge == "" {
		return errors.New("l1 standard bridge is required")
	}

	if c.L2StandardBridge == "" {
		return errors.New("l2 standard bridge is required")
	}

	if c.SlackURL == "" {
		return errors.New("slack url is required")
	}

	if len(c.TokenAddresses) == 0 {
		return errors.New("token addresses is required")
	}

	if c.TonAddress == "" {
		return errors.New("ton address is required")
	}

	return nil
}
