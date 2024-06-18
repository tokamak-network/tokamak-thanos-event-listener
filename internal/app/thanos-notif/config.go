package thanosnotif

type Config struct {
	Network string

	L1Rpc   string
	L1WsRpc string

	L2Rpc   string
	L2WsRpc string

	L1StandardBridge string
	L2StandardBridge string

	L1CrossDomainMessenger string
	L2CrossDomainMessenger string

	L2ToL1MessagePasser string
	OptimismPortal      string

	SlackURL string

	OFF bool
}
