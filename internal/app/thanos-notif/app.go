package thanosnotif

import (
	"context"
	"fmt"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/redis"
	"golang.org/x/sync/errgroup"

	redislib "github.com/go-redis/redis/v8"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/bcclient"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/listener"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/notification"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/repository"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

const (
	ETHDepositInitiatedEventABI      = "ETHDepositInitiated(address,address,uint256,bytes)"
	ETHWithdrawalFinalizedEventABI   = "ETHWithdrawalFinalized(address,address,uint256,bytes)"
	ERC20DepositInitiatedEventABI    = "ERC20DepositInitiated(address,address,address,address,uint256,bytes)"
	ERC20WithdrawalFinalizedEventABI = "ERC20WithdrawalFinalized(address,address,address,address,uint256,bytes)"
	DepositFinalizedEventABI         = "DepositFinalized(address,address,address,address,uint256,bytes)"
	WithdrawalInitiatedEventABI      = "WithdrawalInitiated(address,address,address,address,uint256,bytes)"
)

type App struct {
	cfg          *Config
	l1TokensInfo map[string]*types.Token
	l2TokensInfo map[string]*types.Token
	l1Listener   *listener.EventService
	l2Listener   *listener.EventService
	l1Client     *bcclient.Client
	l2Client     *bcclient.Client
}

func New(ctx context.Context, cfg *Config) (*App, error) {
	redisClient, err := redis.New(ctx, cfg.RedisConfig)
	if err != nil {
		log.GetLogger().Errorw("Failed to connect to redis", "error", err)
		return nil, err
	}

	l1Client, err := bcclient.New(ctx, cfg.L1WsRpc, cfg.L1HttpRpc)
	if err != nil {
		log.GetLogger().Errorw("Failed to create L1 client", "error", err)
		return nil, err
	}

	l2Client, err := bcclient.New(ctx, cfg.L2WsRpc, cfg.L2HttpRpc)
	if err != nil {
		log.GetLogger().Errorw("Failed to create L2 client", "error", err)
		return nil, err
	}

	l1Tokens, err := fetchTokensInfo(l1Client, cfg.L1TokenAddresses)
	if err != nil {
		log.GetLogger().Errorw("Failed to fetch L1 tokens info", "error", err)
		return nil, err
	}

	l2Tokens, err := fetchTokensInfo(l2Client, cfg.L2TokenAddresses)
	if err != nil {
		log.GetLogger().Errorw("Failed to fetch L2 tokens info", "error", err)
		return nil, err
	}

	app := &App{
		cfg:          cfg,
		l1TokensInfo: l1Tokens,
		l2TokensInfo: l2Tokens,
		l1Client:     l1Client,
		l2Client:     l2Client,
	}

	slackNotifier := notification.MakeSlackNotificationService(cfg.SlackURL, 5)

	l1Listener, err := app.initL1Listener(ctx, slackNotifier, l1Client, redisClient)
	if err != nil {
		log.GetLogger().Errorw("Failed to initialize L1 listener", "error", err)
		return nil, err
	}

	l2Listener, err := app.initL2Listener(ctx, slackNotifier, l2Client, redisClient)
	if err != nil {
		log.GetLogger().Errorw("Failed to initialize L2 listener", "error", err)
		return nil, err
	}

	app.l1Listener = l1Listener
	app.l2Listener = l2Listener

	return app, nil
}

func (p *App) Start(ctx context.Context) error {
	var g errgroup.Group

	g.Go(func() error {
		err := p.l1Listener.Start(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err := p.l2Listener.Start(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.GetLogger().Errorw("Failed to start service", "error", err)
		return err
	}

	return nil
}

func (p *App) initL1Listener(ctx context.Context, slackNotifier *notification.SlackNotificationService, l1Client *bcclient.Client, redisClient redislib.UniversalClient) (*listener.EventService, error) {
	l1SyncBlockMetadataRepo := repository.NewSyncBlockMetadataRepository(fmt.Sprintf("%s:%s", p.cfg.Network, "l1"), redisClient)
	l1BlockKeeper, err := repository.NewBlockKeeper(ctx, l1Client, l1SyncBlockMetadataRepo)
	if err != nil {
		log.GetLogger().Errorw("Failed to create L1 block keeper", "error", err)
		return nil, err
	}

	l1Service, err := listener.MakeService("l1-event-listener", l1Client, l1BlockKeeper)
	if err != nil {
		log.GetLogger().Errorw("Failed to make L1 service", "error", err)
		return nil, err
	}

	// L1StandardBridge ETH deposit and withdrawal
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L1StandardBridge, ETHDepositInitiatedEventABI, p.depositETHInitiatedEvent))
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L1StandardBridge, ETHWithdrawalFinalizedEventABI, p.withdrawalETHFinalizedEvent))

	// L1StandardBridge ERC20 deposit and withdrawal
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L1StandardBridge, ERC20DepositInitiatedEventABI, p.depositERC20InitiatedEvent))
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L1StandardBridge, ERC20WithdrawalFinalizedEventABI, p.withdrawalERC20FinalizedEvent))

	// L1UsdcBridge ERC20 deposit and withdrawal
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L1UsdcBridge, ERC20DepositInitiatedEventABI, p.depositUsdcInitiatedEvent))
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L1UsdcBridge, ERC20WithdrawalFinalizedEventABI, p.withdrawalUsdcFinalizedEvent))

	return l1Service, nil
}

func (p *App) initL2Listener(ctx context.Context, slackNotifier *notification.SlackNotificationService, l2Client *bcclient.Client, redisClient redislib.UniversalClient) (*listener.EventService, error) {
	l2SyncBlockMetadataRepo := repository.NewSyncBlockMetadataRepository(fmt.Sprintf("%s:%s", p.cfg.Network, "l2"), redisClient)
	l2BlockKeeper, err := repository.NewBlockKeeper(ctx, l2Client, l2SyncBlockMetadataRepo)
	if err != nil {
		log.GetLogger().Errorw("Failed to make L2 service", "error", err)
		return nil, err
	}

	l2Service, err := listener.MakeService("l2-event-listener", l2Client, l2BlockKeeper)
	if err != nil {
		log.GetLogger().Errorw("Failed to make L2 service", "error", err)
		return nil, err
	}

	// L2StandardBridge deposit and withdrawal
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L2StandardBridge, DepositFinalizedEventABI, p.depositFinalizedEvent))
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L2StandardBridge, WithdrawalInitiatedEventABI, p.withdrawalInitiatedEvent))

	// L2UsdcBridge ERC20 deposit and withdrawal
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L2UsdcBridge, DepositFinalizedEventABI, p.depositUsdcFinalizedEvent))
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(slackNotifier, p.cfg.L2UsdcBridge, WithdrawalInitiatedEventABI, p.withdrawalUsdcInitiatedEvent))

	return l2Service, nil
}
