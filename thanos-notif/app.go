package thanosnotif

import "github.com/tokamak-network/tokamak-thanos-event-listener/notification"

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

func (app *NotifApp) Start() error {
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
