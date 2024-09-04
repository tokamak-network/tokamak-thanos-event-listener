package thanosnotif

import (
	"fmt"

	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tokamak-network/tokamak-thanos/op-bindings/bindings"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/erc20"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

func (p *App) withdrawalETHFinalizedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got ETH Withdrawal Event", "event", vLog)

	l1BridgeFilterer, _, err := p.getBridgeFilterers()
	if err != nil {
		return "", "", err
	}

	event, err := l1BridgeFilterer.ParseETHWithdrawalFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("ETHWithdrawalFinalized event log parsing fail", "error", err)
		return "", "", err
	}

	ethWith := bindings.L1StandardBridgeETHWithdrawalFinalized{
		From:   event.From,
		To:     event.To,
		Amount: event.Amount,
	}

	Amount := formatAmount(ethWith.Amount, 18)

	// Slack notify title and text
	title := fmt.Sprintf("[" + p.cfg.Network + "] [ETH Withdrawal Finalized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nAmount: %s ETH", vLog.TxHash, ethWith.From, ethWith.To, Amount)

	return title, text, nil
}

func (p *App) withdrawalERC20FinalizedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got ERC20 Withdrawal Event", "event", vLog)

	l1BridgeFilterer, _, err := p.getBridgeFilterers()
	if err != nil {
		return "", "", err
	}

	event, err := l1BridgeFilterer.ParseERC20WithdrawalFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("ERC20WithdrawalFinalized event parsing fail", "error", err)
		return "", "", err
	}

	erc20With := bindings.L1StandardBridgeERC20WithdrawalFinalized{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	// get symbol and decimals
	l1Token := erc20With.L1Token
	l1TokenInfo, found := p.l1TokensInfo[l1Token.Hex()]
	if !found {
		newToken, err := erc20.FetchTokenInfo(p.l1Client, l1Token.Hex())
		if err != nil || newToken == nil {
			log.GetLogger().Errorw("Token info not found for address", "l1Token", l1Token.Hex())
			return "", "", err
		}
		l1TokenInfo = newToken
		p.mu.Lock()
		p.l1TokensInfo[l1Token.Hex()] = l1TokenInfo
		p.mu.Unlock()
	}

	tokenSymbol := l1TokenInfo.Symbol
	tokenDecimals := l1TokenInfo.Decimals

	amount := formatAmount(erc20With.Amount, tokenDecimals)

	// Slack notify title and text
	var title string

	isTON := l1TokenInfo.Symbol == "TON"

	if isTON {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [TON Withdrawal Finalized]")
	} else {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ERC-20 Withdrawal Finalized]")
	}
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, erc20With.From, erc20With.To, erc20With.L1Token, erc20With.L2Token, amount, tokenSymbol)

	return title, text, nil
}

func (p *App) withdrawalInitiatedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got L2 Withdrawal Event", "event", vLog)

	_, l2BridgeFilterer, err := p.getBridgeFilterers()
	if err != nil {
		return "", "", err
	}

	event, err := l2BridgeFilterer.ParseWithdrawalInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("WithdrawalInitiated event parsing fail", "error", err)
		return "", "", err
	}

	l2With := bindings.L2StandardBridgeWithdrawalInitiated{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	l2Token := l2With.L2Token

	l2TokenInfo, found := p.l2TokensInfo[l2Token.Hex()]
	if !found {
		newToken, err := erc20.FetchTokenInfo(p.l2Client, l2Token.Hex())
		if err != nil || newToken == nil {
			log.GetLogger().Errorw("Token info not found for address", "l2Token", l2Token.Hex())
			return "", "", err
		}
		l2TokenInfo = newToken
		p.mu.Lock()
		p.l2TokensInfo[l2Token.Hex()] = l2TokenInfo
		p.mu.Unlock()
	}

	if l2TokenInfo == nil {
		return "", "", fmt.Errorf("l2TokenInfo not found")
	}

	tokenSymbol := l2TokenInfo.Symbol
	tokenDecimals := l2TokenInfo.Decimals
	amount := formatAmount(l2With.Amount, tokenDecimals)

	var title string
	var text string

	isETH := l2TokenInfo.Symbol == "ETH"
	isTON := l2TokenInfo.Symbol == "TON"

	if isETH {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ETH Withdrawal Initialized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: ETH\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2With.From, l2With.To, l2With.L2Token, amount, tokenSymbol)
	} else if isTON {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [TON Withdrawal Initialized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2With.From, l2With.To, l2With.L1Token, l2With.L2Token, amount, tokenSymbol)
	} else {
		title = fmt.Sprintf("[" + p.cfg.Network + "] [ERC-20 Withdrawal Initialized]")
		text = fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s %s", vLog.TxHash, l2With.From, l2With.To, l2With.L1Token, l2With.L2Token, amount, tokenSymbol)
	}

	return title, text, nil
}

func (p *App) withdrawalUsdcFinalizedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got L1 USDC Withdrawal Event", "event", vLog)

	l1UsdcBridgeFilterer, _, err := p.getUSDCBridgeFilterers()
	if err != nil {
		return "", "", err
	}

	event, err := l1UsdcBridgeFilterer.ParseERC20WithdrawalFinalized(*vLog)
	if err != nil {
		log.GetLogger().Errorw("USDC WithdrawalFinalized event parsing fail", "error", err)
		return "", "", err
	}

	l1UsdcWith := bindings.L1UsdcBridgeERC20WithdrawalFinalized{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	Amount := formatAmount(l1UsdcWith.Amount, 6)

	// Slack notify title and text
	title := fmt.Sprintf("[" + p.cfg.Network + "] [USDC Withdrawal Finalized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l1UsdcWith.From, l1UsdcWith.To, l1UsdcWith.L1Token, l1UsdcWith.L2Token, Amount)

	return title, text, nil
}

func (p *App) withdrawalUsdcInitiatedEvent(vLog *ethereumTypes.Log) (string, string, error) {
	log.GetLogger().Infow("Got L2 USDC Withdrawal Event", "event", vLog)

	_, l2UsdcBridgeFilterer, err := p.getUSDCBridgeFilterers()
	if err != nil {
		log.GetLogger().Errorw("Failed to get USDC bridge filters", "error", err)
		return "", "", err
	}

	event, err := l2UsdcBridgeFilterer.ParseWithdrawalInitiated(*vLog)
	if err != nil {
		log.GetLogger().Errorw("Failed to parse the USDC WithdrawalInitiated event", "error", err)
		return "", "", err
	}

	l2UsdcWith := bindings.L2UsdcBridgeWithdrawalInitiated{
		L1Token: event.L1Token,
		L2Token: event.L2Token,
		From:    event.From,
		To:      event.To,
		Amount:  event.Amount,
	}

	Amount := formatAmount(l2UsdcWith.Amount, 6)

	title := fmt.Sprintf("[" + p.cfg.Network + "] [USDC Withdrawal Initialized]")
	text := fmt.Sprintf("Tx: "+p.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+p.cfg.L2ExplorerUrl+"/address/%s\nTo: "+p.cfg.L1ExplorerUrl+"/address/%s\nL1Token: "+p.cfg.L1ExplorerUrl+"/token/%s\nL2Token: "+p.cfg.L2ExplorerUrl+"/token/%s\nAmount: %s USDC", vLog.TxHash, l2UsdcWith.From, l2UsdcWith.To, l2UsdcWith.L1Token, l2UsdcWith.L2Token, Amount)

	return title, text, nil
}
