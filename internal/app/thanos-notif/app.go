package thanosnotif

import (
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/listener"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/notification"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

type Notifier interface {
	NotifyWithReTry(title string, text string)
	Notify(title string, text string) error
	Enable()
	Disable()
}

const (
	TransferEventABI = "Transfer(address,address,uint256)"
)

type App struct {
	cfg      *Config
	notifier Notifier
}

func (app *App) ERC20TransferEvent(vLog *types.Log) {
	log.GetLogger().Infow("Got ERC20TransferEvent", "event", vLog)
}

func (app *App) Start() error {
	service := listener.MakeService(app.cfg.L1WsRpc)
	// for testing: listen Transfer event. Replace contract address when you make testing
	for _, transferEventAddress := range app.cfg.TransferEventAddresses {
		depositRelayedRequest := listener.MakeEventRequest(transferEventAddress, TransferEventABI, app.ERC20TransferEvent)
		service.AddSubscribeRequest(depositRelayedRequest)
	}
	err := service.Start()
	if err != nil {
		log.GetLogger().Errorw("Failed to start service", "err", err)
		return err
	}
	return nil
}

func New(config *Config) *App {
	slackNotifSrv := notification.MakeSlackNotificationService(config.SlackURL, 5)

	return &App{
		cfg:      config,
		notifier: slackNotifSrv,
	}
}
