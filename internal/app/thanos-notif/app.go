package thanosnotif

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/redis"
	"golang.org/x/sync/errgroup"

	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	redislib "github.com/go-redis/redis/v8"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/bcclient"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/erc20"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/listener"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/notification"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/repository"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
	"github.com/tokamak-network/tokamak-thanos/op-bindings/bindings"
)

const (
	ETHDepositInitiatedEventABI      = "ETHDepositInitiated(address,address,uint256,bytes)"
	ETHWithdrawalFinalizedEventABI   = "ETHWithdrawalFinalized(address,address,uint256,bytes)"
	ERC20DepositInitiatedEventABI    = "ERC20DepositInitiated(address,address,address,address,uint256,bytes)"
	ERC20WithdrawalFinalizedEventABI = "ERC20WithdrawalFinalized(address,address,address,address,uint256,bytes)"
	DepositFinalizedEventABI         = "DepositFinalized(address,address,address,address,uint256,bytes)"
	WithdrawalInitiatedEventABI      = "WithdrawalInitiated(address,address,address,address,uint256,bytes)"
)

type Notifier interface {
	NotifyWithReTry(title string, text string)
	Notify(title string, text string) error
	Enable()
	Disable()
}

type App struct {
	cfg          *Config
	notifier     Notifier
	tonAddress   string
	l1TokensInfo map[string]*types.Token
	l2TokensInfo map[string]*types.Token
	l1Listener   *listener.EventService
	l2Listener   *listener.EventService
}

func New(ctx context.Context, cfg *Config) (*App, error) {
	slackNotifSrv := notification.MakeSlackNotificationService(cfg.SlackURL, 5)

	redisClient, err := redis.New(ctx, cfg.RedisConfig)
	if err != nil {
		log.GetLogger().Errorw("Failed to connect to redis", "error", err)
		return nil, err
	}

	l1Client, err := bcclient.New(ctx, cfg.L1WsRpc)
	if err != nil {
		log.GetLogger().Errorw("Failed to create L1 client", "error", err)
		return nil, err
	}

	l2Client, err := bcclient.New(ctx, cfg.L2WsRpc)
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
		notifier:     slackNotifSrv,
		tonAddress:   cfg.TonAddress,
		l1TokensInfo: l1Tokens,
		l2TokensInfo: l2Tokens,
	}

	l1Listener, err := app.initL1Listener(ctx, l1Client, redisClient)
	if err != nil {
		log.GetLogger().Errorw("Failed to initialize L1 listener", "error", err)
		return nil, err
	}

	l2Listener, err := app.initL2Listener(ctx, l2Client, redisClient)
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

func (p *App) ETHDepositEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got ETH Deposit Event", "event", vLog)

	l1BridgeFilterer, _, err := p.getBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l1BridgeFilterer.ParseETHDepositInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("ETHDepositInitiated event parsing fail", "error", err)
		return
	}

	ethDep := bindings.L1StandardBridgeETHDepositInitiated{
		From:   event.From,
		To:     event.To,
		Amount: event.Amount,
	}

	Amount := FormatAmount(ethDep.Amount, 18)

	// Slack notify title and text
	title := fmt.Sprintf("[" + p.cfg.Network + "] [ETH Deposit Initialized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nAmount: %s ETH", vLog.TxHash, ethDep.From, ethDep.To, Amount)

	p.notifier.Notify(title, text)
}

func (p *App) ETHWithdrawalEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got ETH Withdrawal Event", "event", vLog)

	l1BridgeFilterer, _, err := p.getBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l1BridgeFilterer.ParseETHWithdrawalFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("ETHWithdrawalFinalized event log parsing fail", "error", err)
		return
	}

	ethWith := bindings.L1StandardBridgeETHWithdrawalFinalized{
		From:   event.From,
		To:     event.To,
		Amount: event.Amount,
	}

	Amount := FormatAmount(ethWith.Amount, 18)

	// Slack notify title and text
	title := fmt.Sprintf("[" + p.cfg.Network + "] [ETH Withdrawal Finalized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nAmount: %s ETH", vLog.TxHash, ethWith.From, ethWith.To, Amount)

	if err := p.notifier.Notify(title, text); err != nil {
		log.GetLogger().Errorw("Failed to notify ETH Event", "error", err)
	}
}

func (p *App) ERC20DepositEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got ERC20 Deposit Event", "event", vLog)

	l1BridgeFilterer, _, err := p.getBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l1BridgeFilterer.ParseERC20DepositInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("ERC20DepositInitiated event parsing fail", "error", err)
		return
	}

	erc20Dep := bindings.L1StandardBridgeERC20DepositInitiated{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	// get symbol and decimals
	tokenAddress := erc20Dep.L1Token
	tokenInfo, found := p.l1TokensInfo[tokenAddress.Hex()]
	if !found {
		log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddress.Hex())
		return
	}

	tokenSymbol := tokenInfo.Symbol
	tokenDecimals := tokenInfo.Decimals

	Amount := FormatAmount(erc20Dep.Amount, tokenDecimals)

	// Slack notify title and text
	var title string

	isTON := tokenAddress.Cmp(common.HexToAddress(p.tonAddress)) == 0

	if isTON {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [TON Deposit Initialized]")
	} else {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ERC-20 Deposit Initialized]")
	}
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, erc20Dep.From, erc20Dep.To, erc20Dep.L1Token, erc20Dep.L2Token, Amount, tokenSymbol)

	p.notifier.Notify(title, text)
}

func (p *App) ERC20WithdrawalEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got ERC20 Withdrawal Event", "event", vLog)

	l1BridgeFilterer, _, err := p.getBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l1BridgeFilterer.ParseERC20WithdrawalFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("ERC20WithdrawalFinalized event parsing fail", "error", err)
		return
	}

	erc20With := bindings.L1StandardBridgeERC20WithdrawalFinalized{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	// get symbol and decimals
	tokenAddress := erc20With.L1Token
	tokenInfo, found := p.l1TokensInfo[tokenAddress.Hex()]
	if !found {
		log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddress.Hex())
		return
	}

	tokenSymbol := tokenInfo.Symbol
	tokenDecimals := tokenInfo.Decimals

	Amount := FormatAmount(erc20With.Amount, tokenDecimals)

	// Slack notify title and text
	var title string

	isTON := tokenAddress.Cmp(common.HexToAddress(p.tonAddress)) == 0

	if isTON {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [TON Withdrawal Finalized]")
	} else {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ERC-20 Withdrawal Finalized]")
	}
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, erc20With.From, erc20With.To, erc20With.L1Token, erc20With.L2Token, Amount, tokenSymbol)

	p.notifier.Notify(title, text)
}

func (p *App) L2DepositEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L2 Deposit Event", "event", vLog)

	_, l2BridgeFilterer, err := p.getBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l2BridgeFilterer.ParseDepositFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("DepositFinalized event parsing fail", "error", err)
		return
	}

	l2Dep := bindings.L2StandardBridgeDepositFinalized{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	// get symbol and decimals
	var tokenSymbol string
	var tokenDecimals int

	tokenAddress := l2Dep.L2Token
	isETH := tokenAddress.Cmp(common.HexToAddress("0x4200000000000000000000000000000000000486")) == 0
	isTON := tokenAddress.Cmp(common.HexToAddress("0xDeadDeAddeAddEAddeadDEaDDEAdDeaDDeAD0000")) == 0

	if isETH {
		tokenSymbol = "ETH"
		tokenDecimals = 18
	} else if isTON {
		tokenSymbol = "TON"
		tokenDecimals = 18
	} else {
		tokenInfo, found := p.l1TokensInfo[tokenAddress.Hex()]
		if !found {
			log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddress.Hex())
			return
		}
		tokenSymbol = tokenInfo.Symbol
		tokenDecimals = tokenInfo.Decimals
	}

	Amount := FormatAmount(l2Dep.Amount, tokenDecimals)

	var title string
	var text string

	if isETH {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ETH Deposit Finalized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: ETH\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2Dep.From, l2Dep.To, l2Dep.L2Token, Amount, tokenSymbol)
	} else if isTON {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [TON Deposit Finalized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2Dep.From, l2Dep.To, p.tonAddress, l2Dep.L2Token, Amount, tokenSymbol)
	} else {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ERC-20 Deposit Finalized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2Dep.From, l2Dep.To, l2Dep.L1Token, l2Dep.L2Token, Amount, tokenSymbol)
	}

	p.notifier.Notify(title, text)
}

func (p *App) L2WithdrawalEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L2 Withdrawal Event", "event", vLog)

	_, l2BridgeFilterer, err := p.getBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l2BridgeFilterer.ParseWithdrawalInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("WithdrawalInitiated event parsing fail", "error", err)
		return
	}

	l2With := bindings.L2StandardBridgeWithdrawalInitiated{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	// get symbol and decimals
	var tokenSymbol string
	var tokenDecimals int

	tokenAddress := l2With.L2Token
	isETH := tokenAddress.Cmp(common.HexToAddress("0x4200000000000000000000000000000000000486")) == 0
	isTON := tokenAddress.Cmp(common.HexToAddress("0xDeadDeAddeAddEAddeadDEaDDEAdDeaDDeAD0000")) == 0

	if isETH {
		tokenSymbol = "ETH"
		tokenDecimals = 18
	} else if isTON {
		tokenSymbol = "TON"
		tokenDecimals = 18
	} else {
		tokenInfo, found := p.l1TokensInfo[tokenAddress.Hex()]
		if !found {
			log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddress.Hex())
			return
		}
		tokenSymbol = tokenInfo.Symbol
		tokenDecimals = tokenInfo.Decimals
	}

	Amount := FormatAmount(l2With.Amount, tokenDecimals)

	var title string
	var text string

	if isETH {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ETH Withdrawal Initialized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: ETH\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2With.From, l2With.To, l2With.L2Token, Amount, tokenSymbol)
	} else if isTON {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [TON Withdrawal Initialized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2With.From, l2With.To, p.tonAddress, l2With.L2Token, Amount, tokenSymbol)
	} else {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ERC-20 Withdrawal Initialized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2With.From, l2With.To, l2With.L1Token, l2With.L2Token, Amount, tokenSymbol)
	}

	p.notifier.Notify(title, text)
}

func (p *App) L1UsdcDepEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L1 USDC Deposit Event", "event", vLog)

	l1UsdcBridgeFilterer, _, err := p.getUsdcBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l1UsdcBridgeFilterer.ParseERC20DepositInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("USDC DepositInitiated event parsing fail", "error", err)
		return
	}

	l1UsdcDep := bindings.L1UsdcBridgeERC20DepositInitiated{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	Amount := FormatAmount(l1UsdcDep.Amount, 6)

	// Slack notify title and text
	title := fmt.Sprintf("[" + p.cfg.Network + "] [USDC Deposit Initialized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l1UsdcDep.From, l1UsdcDep.To, l1UsdcDep.L1Token, l1UsdcDep.L2Token, Amount)

	p.notifier.Notify(title, text)
}

func (p *App) L1UsdcWithEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L1 USDC Withdrawal Event", "event", vLog)

	l1UsdcBridgeFilterer, _, err := p.getUsdcBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l1UsdcBridgeFilterer.ParseERC20WithdrawalFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("USDC WithdrawalFinalized event parsing fail", "error", err)
		return
	}

	l1UsdcWith := bindings.L1UsdcBridgeERC20WithdrawalFinalized{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	Amount := FormatAmount(l1UsdcWith.Amount, 6)

	// Slack notify title and text
	title := fmt.Sprintf("[" + p.cfg.Network + "] [USDC Withdrawal Finalized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l1UsdcWith.From, l1UsdcWith.To, l1UsdcWith.L1Token, l1UsdcWith.L2Token, Amount)

	p.notifier.Notify(title, text)
}

func (p *App) L2UsdcDepEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L2 USDC Deposit Event", "event", vLog)

	_, l2UsdcBridgeFilterer, err := p.getUsdcBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l2UsdcBridgeFilterer.ParseDepositFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("USDC DepositFinalized event parsing fail", "error", err)
		return
	}

	l2UsdcDep := bindings.L2UsdcBridgeDepositFinalized{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	Amount := FormatAmount(l2UsdcDep.Amount, 6)

	title := fmt.Sprintf("[" + p.cfg.Network + "] [USDC Deposit Finalized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l2UsdcDep.From, l2UsdcDep.To, l2UsdcDep.L1Token, l2UsdcDep.L2Token, Amount)

	p.notifier.Notify(title, text)
}

func (p *App) L2UsdcWithEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L2 USDC Withdrawal Event", "event", vLog)

	_, l2UsdcBridgeFilterer, err := p.getUsdcBridgeFilterers()
	if err != nil {
		log.GetLogger().Errorw("Failed to get USDC bridge filters", "error", err)
		return
	}

	event, err := l2UsdcBridgeFilterer.ParseWithdrawalInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("Failed to parse the USDC WithdrawalInitiated event", "error", err)
		return
	}

	l2UsdcWith := bindings.L2UsdcBridgeWithdrawalInitiated{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	Amount := FormatAmount(l2UsdcWith.Amount, 6)

	title := fmt.Sprintf("[" + p.cfg.Network + "] [USDC Withdrawal Initialized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l2UsdcWith.From, l2UsdcWith.To, l2UsdcWith.L1Token, l2UsdcWith.L2Token, Amount)

	err = p.notifier.Notify(title, text)
	if err != nil {
		return
	}
}

func (p *App) initL1Listener(ctx context.Context, l1Client *bcclient.Client, redisClient redislib.UniversalClient) (*listener.EventService, error) {
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
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L1StandardBridge, ETHDepositInitiatedEventABI, p.ETHDepositEvent))
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L1StandardBridge, ETHWithdrawalFinalizedEventABI, p.ETHWithdrawalEvent))

	// L1StandardBridge ERC20 deposit and withdrawal
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L1StandardBridge, ERC20DepositInitiatedEventABI, p.ERC20DepositEvent))
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L1StandardBridge, ERC20WithdrawalFinalizedEventABI, p.ERC20WithdrawalEvent))

	// L1UsdcBridge ERC20 deposit and withdrawal
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L1UsdcBridge, ERC20DepositInitiatedEventABI, p.L1UsdcDepEvent))
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L1UsdcBridge, ERC20WithdrawalFinalizedEventABI, p.L1UsdcWithEvent))

	return l1Service, nil
}

func (p *App) initL2Listener(ctx context.Context, l2Client *bcclient.Client, redisClient redislib.UniversalClient) (*listener.EventService, error) {
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
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L2StandardBridge, DepositFinalizedEventABI, p.L2DepositEvent))
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L2StandardBridge, WithdrawalInitiatedEventABI, p.L2WithdrawalEvent))

	// L2UsdcBridge ERC20 deposit and withdrawal
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L2UsdcBridge, DepositFinalizedEventABI, p.L2UsdcDepEvent))
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(p.cfg.L2UsdcBridge, WithdrawalInitiatedEventABI, p.L2UsdcWithEvent))

	return l2Service, nil
}

func fetchTokensInfo(bcClient *bcclient.Client, tokenAddresses []string) (map[string]*types.Token, error) {
	tokenInfoMap := make(map[string]*types.Token)
	for _, tokenAddress := range tokenAddresses {
		tokenInfo, err := erc20.FetchTokenInfo(bcClient, tokenAddress)
		if err != nil {
			log.GetLogger().Errorw("Failed to fetch token info", "error", err, "address", tokenAddress)
			return nil, err
		}

		if tokenInfo == nil {
			log.GetLogger().Errorw("Token info empty", "address", tokenAddress)
			return nil, errors.New("token info is empty")
		}

		log.GetLogger().Infow("Got token info", "token", tokenInfo)

		tokenInfoMap[tokenAddress] = tokenInfo
	}

	return tokenInfoMap, nil
}
