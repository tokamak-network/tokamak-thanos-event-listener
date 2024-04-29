package thanosnotif

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	corelistener "github.com/tokamak-network/tokamak-thanos-event-listener/core-listener"
	"github.com/tokamak-network/tokamak-thanos-event-listener/notification"
)

type NotifAppConfig struct {
	L1_RPC string
	L2_RPC string

	L1StandBridge string
	L2StandBridge string

	L1CrossDomainMessenger string
	L2CrossDomainMessenger string

	L2ToL1MessagePasser string
	OptimismPortal      string

	SlackURL string
}

type NotifApp struct {
	config       *NotifAppConfig
	notifService notification.INotifService
}

func (app *NotifApp) ERC20TransferEvent(vLog *types.Log) {
	fmt.Println("ERC20TransferEvent: ", vLog)
}

func (app *NotifApp) Start() error {
	service := corelistener.MakeService("ws://localhost:8546")
	// for testing: listen Transfer event. Replace contract address when you make testing
	depositRelayedRequest := corelistener.MakeEventRequest("0xC7844340d14deAedfDD2f2dD9360c336661b2F0A", "Transfer(address,address,uint256)", app.ERC20TransferEvent)
	service.AddSubscribeRequest(depositRelayedRequest)
	service.Start()
	time.Sleep(time.Second)
	return nil
}

func (app *NotifApp) initialize() {
	app.notifService = notification.MakeSlackNotificationService(app.config.SlackURL, 5)
}

func MakeNotifApp(config *NotifAppConfig) *NotifApp {
	app := &NotifApp{config: config}
	app.initialize()
	return app
}
