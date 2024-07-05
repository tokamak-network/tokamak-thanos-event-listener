package thanosnotif

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
}
