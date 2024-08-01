package thanosnotif

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tokamak-network/tokamak-thanos/op-bindings/bindings"
	"golang.org/x/sync/errgroup"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/erc20"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/listener"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/notification"
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

type Notifier interface {
	NotifyWithReTry(title string, text string)
	Notify(title string, text string) error
	Enable()
	Disable()
}

type App struct {
	cfg        *Config
	notifier   Notifier
	tonAddress string
	tokenInfo  map[string]*types.Token
	ethClient  *ethclient.Client
}

func New(config *Config) (*App, error) {
	slackNotifSrv := notification.MakeSlackNotificationService(config.SlackURL, 5)
	ethClient, err := ethclient.Dial(config.L1Rpc)
	if err != nil {
		return nil, err
	}
	app := &App{
		cfg:        config,
		notifier:   slackNotifSrv,
		tokenInfo:  make(map[string]*types.Token),
		ethClient:  ethClient,
		tonAddress: config.TonAddress,
	}

	err = app.fetchTokenInfo()
	if err != nil {
		log.GetLogger().Errorw("Failed to update token info", "error", err)
		return nil, err
	}

	return app, nil
}

func (app *App) ETHDepEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got ETH Deposit Event", "event", vLog)

	l1BridgeFilterer, _, err := app.getBridgeFilterers()
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

	Amount := app.formatAmount(ethDep.Amount, 18)

	// Slack notify title and text
	title := fmt.Sprintf("[" + app.cfg.Network + "] [ETH Deposit Initialized]")
	text := fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nAmount: %s ETH", vLog.TxHash, ethDep.From, ethDep.To, Amount)

	app.notifier.Notify(title, text)
}

func (app *App) ETHWithEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got ETH Withdrawal Event", "event", vLog)

	l1BridgeFilterer, _, err := app.getBridgeFilterers()
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

	Amount := app.formatAmount(ethWith.Amount, 18)

	// Slack notify title and text
	title := fmt.Sprintf("[" + app.cfg.Network + "] [ETH Withdrawal Finalized]")
	text := fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nAmount: %s ETH", vLog.TxHash, ethWith.From, ethWith.To, Amount)

	if err := app.notifier.Notify(title, text); err != nil {
		log.GetLogger().Errorw("Failed to notify ETH Event", "error", err)
	}
}

func (app *App) ERC20DepEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got ERC20 Deposit Event", "event", vLog)

	l1BridgeFilterer, _, err := app.getBridgeFilterers()
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
	tokenInfo, found := app.tokenInfo[tokenAddress.Hex()]
	if !found {
		log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddress.Hex())
		return
	}

	tokenSymbol := tokenInfo.Symbol
	tokenDecimals := tokenInfo.Decimals

	Amount := app.formatAmount(erc20Dep.Amount, tokenDecimals)

	// Slack notify title and text
	var title string

	isTON := tokenAddress.Cmp(common.HexToAddress(app.tonAddress)) == 0

	if isTON {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [TON Deposit Initialized]")
	} else {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ERC-20 Deposit Initialized]")
	}
	text := fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, erc20Dep.From, erc20Dep.To, erc20Dep.L1Token, erc20Dep.L2Token, Amount, tokenSymbol)

	app.notifier.Notify(title, text)
}

func (app *App) ERC20WithEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got ERC20 Withdrawal Event", "event", vLog)

	l1BridgeFilterer, _, err := app.getBridgeFilterers()
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
	tokenInfo, found := app.tokenInfo[tokenAddress.Hex()]
	if !found {
		log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddress.Hex())
		return
	}

	tokenSymbol := tokenInfo.Symbol
	tokenDecimals := tokenInfo.Decimals

	Amount := app.formatAmount(erc20With.Amount, tokenDecimals)

	// Slack notify title and text
	var title string

	isTON := tokenAddress.Cmp(common.HexToAddress(app.tonAddress)) == 0

	if isTON {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [TON Withdrawal Finalized]")
	} else {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ERC-20 Withdrawal Finalized]")
	}
	text := fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, erc20With.From, erc20With.To, erc20With.L1Token, erc20With.L2Token, Amount, tokenSymbol)

	app.notifier.Notify(title, text)
}

func (app *App) L2DepEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L2 Deposit Event", "event", vLog)

	_, l2BridgeFilterer, err := app.getBridgeFilterers()
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
		tokenInfo, found := app.tokenInfo[tokenAddress.Hex()]
		if !found {
			log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddress.Hex())
			return
		}
		tokenSymbol = tokenInfo.Symbol
		tokenDecimals = tokenInfo.Decimals
	}

	Amount := app.formatAmount(l2Dep.Amount, tokenDecimals)

	var title string
	var text string

	if isETH {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ETH Deposit Finalized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nL1Token: ETH\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2Dep.From, l2Dep.To, l2Dep.L2Token, Amount, tokenSymbol)
	} else if isTON {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [TON Deposit Finalized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2Dep.From, l2Dep.To, app.tonAddress, l2Dep.L2Token, Amount, tokenSymbol)
	} else {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ERC-20 Deposit Finalized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2Dep.From, l2Dep.To, l2Dep.L1Token, l2Dep.L2Token, Amount, tokenSymbol)
	}

	app.notifier.Notify(title, text)
}

func (app *App) L2WithEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L2 Withdrawal Event", "event", vLog)

	_, l2BridgeFilterer, err := app.getBridgeFilterers()
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
		tokenInfo, found := app.tokenInfo[tokenAddress.Hex()]
		if !found {
			log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddress.Hex())
			return
		}
		tokenSymbol = tokenInfo.Symbol
		tokenDecimals = tokenInfo.Decimals
	}

	Amount := app.formatAmount(l2With.Amount, tokenDecimals)

	var title string
	var text string

	if isETH {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ETH Withdrawal Initialized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nL1Token: ETH\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2With.From, l2With.To, l2With.L2Token, Amount, tokenSymbol)
	} else if isTON {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [TON Withdrawal Initialized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2With.From, l2With.To, app.tonAddress, l2With.L2Token, Amount, tokenSymbol)
	} else {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ERC-20 Withdrawal Initialized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2With.From, l2With.To, l2With.L1Token, l2With.L2Token, Amount, tokenSymbol)
	}

	app.notifier.Notify(title, text)
}

func (app *App) L1UsdcDepEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L1 USDC Deposit Event", "event", vLog)

	l1UsdcBridgeFilterer, _, err := app.getUsdcBridgeFilterers()
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

	Amount := app.formatAmount(l1UsdcDep.Amount, 6)

	// Slack notify title and text
	title := fmt.Sprintf("[" + app.cfg.Network + "] [USDC Deposit Initialized]")
	text := fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l1UsdcDep.From, l1UsdcDep.To, l1UsdcDep.L1Token, l1UsdcDep.L2Token, Amount)

	app.notifier.Notify(title, text)
}

func (app *App) L1UsdcWithEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L1 USDC Withdrawal Event", "event", vLog)

	l1UsdcBridgeFilterer, _, err := app.getUsdcBridgeFilterers()
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

	Amount := app.formatAmount(l1UsdcWith.Amount, 6)

	// Slack notify title and text
	title := fmt.Sprintf("[" + app.cfg.Network + "] [USDC Withdrawal Finalized]")
	text := fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l1UsdcWith.From, l1UsdcWith.To, l1UsdcWith.L1Token, l1UsdcWith.L2Token, Amount)

	app.notifier.Notify(title, text)
}

func (app *App) L2UsdcDepEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L2 USDC Deposit Event", "event", vLog)

	_, l2UsdcBridgeFilterer, err := app.getUsdcBridgeFilterers()
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

	Amount := app.formatAmount(l2UsdcDep.Amount, 6)

	title := fmt.Sprintf("[" + app.cfg.Network + "] [USDC Deposit Finalized]")
	text := fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l2UsdcDep.From, l2UsdcDep.To, l2UsdcDep.L1Token, l2UsdcDep.L2Token, Amount)

	app.notifier.Notify(title, text)
}

func (app *App) L2UsdcWithEvent(vLog *ethereumTypes.Log) {
	log.GetLogger().Infow("Got L2 USDC Withdrawal Event", "event", vLog)

	_, l2UsdcBridgeFilterer, err := app.getUsdcBridgeFilterers()
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

	Amount := app.formatAmount(l2UsdcWith.Amount, 6)

	title := fmt.Sprintf("[" + app.cfg.Network + "] [USDC Withdrawal Initialized]")
	text := fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l2UsdcWith.From, l2UsdcWith.To, l2UsdcWith.L1Token, l2UsdcWith.L2Token, Amount)

	err = app.notifier.Notify(title, text)
	if err != nil {
		return
	}
}

func (app *App) Start() error {
	l1Service, err := listener.MakeService(app.cfg.L1WsRpc)
	if err != nil {
		log.GetLogger().Errorw("Failed to make L1 service", "error", err)
		return err
	}

	l2Service, err := listener.MakeService(app.cfg.L2WsRpc)
	if err != nil {
		log.GetLogger().Errorw("Failed to make L2 service", "error", err)
		return err
	}

	// L1StandardBridge ETH deposit and withdrawal
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L1StandardBridge, ETHDepositInitiatedEventABI, app.ETHDepEvent))
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L1StandardBridge, ETHWithdrawalFinalizedEventABI, app.ETHWithEvent))

	// L1StandardBridge ERC20 deposit and withdrawal
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L1StandardBridge, ERC20DepositInitiatedEventABI, app.ERC20DepEvent))
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L1StandardBridge, ERC20WithdrawalFinalizedEventABI, app.ERC20WithEvent))

	// L2StandardBridge deposit and withdrawal
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L2StandardBridge, DepositFinalizedEventABI, app.L2DepEvent))
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L2StandardBridge, WithdrawalInitiatedEventABI, app.L2WithEvent))

	// L1UsdcBridge ERC20 deposit and withdrawal
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L1UsdcBridge, ERC20DepositInitiatedEventABI, app.L1UsdcDepEvent))
	l1Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L1UsdcBridge, ERC20WithdrawalFinalizedEventABI, app.L1UsdcWithEvent))

	// L2UsdcBridge ERC20 deposit and withdrawal
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L2UsdcBridge, DepositFinalizedEventABI, app.L2UsdcDepEvent))
	l2Service.AddSubscribeRequest(listener.MakeEventRequest(app.cfg.L2UsdcBridge, WithdrawalInitiatedEventABI, app.L2UsdcWithEvent))

	var g errgroup.Group

	g.Go(func() error {
		err := l1Service.Start()
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err := l2Service.Start()
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

func (app *App) fetchTokenInfo() error {
	tokenAddresses := app.cfg.TokenAddresses
	tokenInfoMap := make(map[string]*types.Token)
	for _, tokenAddress := range tokenAddresses {
		tokenInfo, err := erc20.FetchTokenInfo(app.ethClient, tokenAddress)
		if err != nil {
			log.GetLogger().Errorw("Failed to fetch token info", "error", err, "address", tokenAddress)
			return err
		}

		if tokenInfo == nil {
			log.GetLogger().Errorw("Token info empty", "address", tokenAddress)
			return errors.New("token info is empty")
		}

		log.GetLogger().Infow("Got token info", "token", tokenInfo)

		tokenInfoMap[tokenAddress] = tokenInfo
	}

	app.tokenInfo = tokenInfoMap

	return nil
}
