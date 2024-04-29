package thanosnotif

type Config struct {
	L1Rpc   string
	L1WsRpc string

	L2Rpc   string
	L2WsRpc string

	L1StandBridge string
	L2StandBridge string

	L1CrossDomainMessenger string
	L2CrossDomainMessenger string

	L2ToL1MessagePasser string
	OptimismPortal      string

	SlackURL string

	TransferEventAddresses []string
}
