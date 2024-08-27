package thanosnotif

import (
	"fmt"

	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
	"github.com/tokamak-network/tokamak-thanos/op-bindings/bindings"
)

func (p *App) depositETHInitiatedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got ETH Deposit Event", "event", vLog)

	l1BridgeFilterer, _, err := p.getBridgeFilterers()
	if err != nil {
		return "", "", err
	}

	event, err := l1BridgeFilterer.ParseETHDepositInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("ETHDepositInitiated event parsing fail", "error", err)
		return "", "", err
	}

	ethDep := bindings.L1StandardBridgeETHDepositInitiated{
		From:   event.From,
		To:     event.To,
		Amount: event.Amount,
	}

	Amount := formatAmount(ethDep.Amount, 18)

	// Slack notify title and text
	title := fmt.Sprintf("[" + p.cfg.Network + "] [ETH Deposit Initialized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nAmount: %s ETH", vLog.TxHash, ethDep.From, ethDep.To, Amount)

	return title, text, nil
}

func (p *App) depositERC20InitiatedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got ERC20 Deposit Event", "event", vLog)

	l1BridgeFilterer, _, err := p.getBridgeFilterers()
	if err != nil {
		return "", "", err
	}

	event, err := l1BridgeFilterer.ParseERC20DepositInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("ERC20DepositInitiated event parsing fail", "error", err)
		return "", "", err
	}

	erc20Dep := bindings.L1StandardBridgeERC20DepositInitiated{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	// get symbol and decimals
	l1Token := erc20Dep.L1Token
	l1TokenInfo, found := p.l1TokensInfo[l1Token.Hex()]
	if !found {
		log.GetLogger().Errorw("Token info not found for address", "l1Token", l1Token.Hex())
		return "", "", err
	}

	tokenSymbol := l1TokenInfo.Symbol
	tokenDecimals := l1TokenInfo.Decimals

	amount := formatAmount(erc20Dep.Amount, tokenDecimals)

	// Slack notify title and text
	var title string

	isTON := l1TokenInfo.Symbol == "TON"

	if isTON {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [TON Deposit Initialized]")
	} else {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ERC-20 Deposit Initialized]")
	}
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, erc20Dep.From, erc20Dep.To, erc20Dep.L1Token, erc20Dep.L2Token, amount, tokenSymbol)

	return title, text, nil
}

func (p *App) depositFinalizedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got L2 Deposit Event", "event", vLog)

	_, l2BridgeFilterer, err := p.getBridgeFilterers()
	if err != nil {
		return "", "", err
	}

	event, err := l2BridgeFilterer.ParseDepositFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("DepositFinalized event parsing fail", "error", err)
		return "", "", err
	}

	l2Dep := bindings.L2StandardBridgeDepositFinalized{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	// get symbol and decimals
	l2Token := l2Dep.L2Token

	l2TokenInfo, found := p.l2TokensInfo[l2Token.Hex()]
	if !found {
		log.GetLogger().Errorw("Token info not found for address", "l2Token", l2Token.Hex())
		return "", "", err
	}

	amount := formatAmount(l2Dep.Amount, l2TokenInfo.Decimals)

	var title string
	var text string

	isETH := l2TokenInfo.Symbol == "ETH"
	isTON := l2TokenInfo.Symbol == "TON"

	if isETH {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ETH Deposit Finalized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: ETH\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2Dep.From, l2Dep.To, l2Dep.L2Token, amount, l2TokenInfo.Symbol)
	} else if isTON {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [TON Deposit Finalized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2Dep.From, l2Dep.To, l2Dep.L1Token, l2Dep.L2Token, amount, l2TokenInfo.Symbol)
	} else {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ERC-20 Deposit Finalized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2Dep.From, l2Dep.To, l2Dep.L1Token, l2Dep.L2Token, amount, l2TokenInfo.Symbol)
	}

	return title, text, nil
}

func (p *App) depositUsdcInitiatedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got L1 USDC Deposit Event", "event", vLog)

	l1UsdcBridgeFilterer, _, err := p.getUSDCBridgeFilterers()
	if err != nil {
		return "", "", err
	}

	event, err := l1UsdcBridgeFilterer.ParseERC20DepositInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("USDC DepositInitiated event parsing fail", "error", err)
		return "", "", err
	}

	l1UsdcDep := bindings.L1UsdcBridgeERC20DepositInitiated{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	amount := formatAmount(l1UsdcDep.Amount, 6)

	// Slack notify title and text
	title := fmt.Sprintf("[" + p.cfg.Network + "] [USDC Deposit Initialized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\namount: %s USDC", vLog.TxHash, l1UsdcDep.From, l1UsdcDep.To, l1UsdcDep.L1Token, l1UsdcDep.L2Token, amount)

	return title, text, nil
}

func (p *App) depositUsdcFinalizedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got L2 USDC Deposit Event", "event", vLog)

	_, l2UsdcBridgeFilterer, err := p.getUSDCBridgeFilterers()
	if err != nil {
		return "", "", err
	}

	event, err := l2UsdcBridgeFilterer.ParseDepositFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("USDC DepositFinalized event parsing fail", "error", err)
		return "", "", err
	}

	l2UsdcDep := bindings.L2UsdcBridgeDepositFinalized{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	amount := formatAmount(l2UsdcDep.Amount, 6)

	title := fmt.Sprintf("[" + p.cfg.Network + "] [USDC Deposit Finalized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L1ExplorerUrl+"/address/%s\nTo: "+p.cfg.L2ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l2UsdcDep.From, l2UsdcDep.To, l2UsdcDep.L1Token, l2UsdcDep.L2Token, amount)

	return title, text, nil
}
