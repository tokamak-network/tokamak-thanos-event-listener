package thanosnotif

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tokamak-network/tokamak-thanos/op-bindings/bindings"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/listener"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/notification"
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
	tokenInfo  map[string]TokenInfo
}

func (app *App) ETHDepEvent(vLog *types.Log) {
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

func (app *App) ETHWithEvent(vLog *types.Log) {
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

	app.notifier.Notify(title, text)
}

func (app *App) ERC20DepEvent(vLog *types.Log) {
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

func (app *App) ERC20WithEvent(vLog *types.Log) {
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

func (app *App) L2DepEvent(vLog *types.Log) {
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

func (app *App) L2WithEvent(vLog *types.Log) {
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

func (app *App) L1UsdcDepEvent(vLog *types.Log) {
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

func (app *App) L1UsdcWithEvent(vLog *types.Log) {
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

func (app *App) L2UsdcDepEvent(vLog *types.Log) {
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

func (app *App) L2UsdcWithEvent(vLog *types.Log) {
	log.GetLogger().Infow("Got L2 USDC Withdrawal Event", "event", vLog)

	_, l2UsdcBridgeFilterer, err := app.getUsdcBridgeFilterers()
	if err != nil {
		return
	}

	event, err := l2UsdcBridgeFilterer.ParseWithdrawalInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("USDC WithdrawalInitiated event parsing fail", "error", err)
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

	app.notifier.Notify(title, text)
}

func (app *App) Start() error {
	l1Service := listener.MakeService(app.cfg.L1WsRpc)
	l2Service := listener.MakeService(app.cfg.L2WsRpc)

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

	err := app.updateTokenInfo()
	if err != nil {
		log.GetLogger().Errorw("Failed to update token info", "err", err)
		return err
	}

	// Start both services
	errCh := make(chan error, 2)

	go func() {
		errCh <- l1Service.Start()
	}()

	go func() {
		errCh <- l2Service.Start()
	}()

	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			log.GetLogger().Errorw("Failed to start service", "err", err)
			return err
		}
	}

	return nil
}

func (app *App) updateTokenInfo() error {
	data := &Data{cfg: app.cfg}
	tokenInfoMap, err := data.tokenInfoMap()
	if err != nil {
		return err
	}

	for k, v := range tokenInfoMap {
		if v.Symbol == "TON" {
			app.tonAddress = k
			break
		}
	}

	app.tokenInfo = tokenInfoMap

	return nil
}

func New(config *Config) *App {
	slackNotifSrv := notification.MakeSlackNotificationService(config.SlackURL, 5)

	app := &App{
		cfg:       config,
		notifier:  slackNotifSrv,
		tokenInfo: make(map[string]TokenInfo),
	}

	return app
}
