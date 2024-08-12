package thanosnotif

import (
	"errors"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/redis"
)

type Config struct {
	Network string

	L1HttpRpc string
	L1WsRpc   string

	L2HttpRpc string
	L2WsRpc   string

	L1StandardBridge string
	L2StandardBridge string

	L1UsdcBridge string
	L2UsdcBridge string

	SlackURL string

	L1ExplorerUrl string
	L2ExplorerUrl string

	L1TokenAddresses []string
	L2TokenAddresses []string

	RedisConfig redis.Config
}

func (c *Config) Validate() error {
	if c.L1WsRpc == "" {
		return errors.New("l1 ws rpc address is required")
	}

	if c.L1HttpRpc == "" {
		return errors.New("l1 http rpc address is required")
	}

	if c.L2WsRpc == "" {
		return errors.New("l2 ws rpc address is required")
	}

	if c.L2HttpRpc == "" {
		return errors.New("l2 http rpc address is required")
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

	if len(c.L1TokenAddresses) == 0 {
		return errors.New("token addresses is required")
	}

	if c.RedisConfig.Addresses == "" {
		return errors.New("redis address is required")
	}

	return nil
}
