package thanosnotif

type Config struct {
	L1_RPC    string
	L1_WS_RPC string

	L2_RPC    string
	L2_WS_RPC string

	L1StandBridge string
	L2StandBridge string

	L1CrossDomainMessenger string
	L2CrossDomainMessenger string

	L2ToL1MessagePasser string
	OptimismPortal      string

	SlackURL string

	TransferEventAddresses []string
}
