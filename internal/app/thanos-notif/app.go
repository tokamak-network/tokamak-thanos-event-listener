package thanosnotif

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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
	ETHDepositInitiatedEventABI      = "ETHDepositInitiated(address,address,uint256,bytes)"
	ETHWithdrawalFinalizedEventABI   = "ETHWithdrawalFinalized(address,address,uint256,bytes)"
	ERC20DepositInitiatedEventABI    = "ERC20DepositInitiated(address,address,address,address,uint256,bytes)"
	ERC20WithdrawalFinalizedEventABI = "ERC20WithdrawalFinalized(address,address,address,address,uint256,bytes)"
	DepositFinalizedEventABI         = "DepositFinalized(address,address,address,address,uint256,bytes)"
	WithdrawalInitiatedEventABI      = "WithdrawalInitiated(address,address,address,address,uint256,bytes)"
)

type App struct {
	cfg       *Config
	notifier  Notifier
	tokenInfo map[string]TokenInfo
}

func (app *App) ETHDepAndWithEvent(vLog *types.Log) {
	log.GetLogger().Infow("Got ETH Deposit or Withdrawal Event", "event", vLog)

	// check the length vLog.Topics and vLog.Data
	if len(vLog.Topics) > 3 {
		log.GetLogger().Errorw("Error: Length of vLog.Topics is not as expected.")
		return
	}

	if len(vLog.Data) < 32 {
		log.GetLogger().Errorw("Error: Length of vLog.Data is not as expected.")
		return
	}

	txHash := vLog.TxHash
	From := common.HexToAddress(vLog.Topics[1].Hex())
	To := common.HexToAddress(vLog.Topics[2].Hex())

	// ETH deposit and withdrawal Amount
	amountData := vLog.Data[:32]

	decimals := 18
	value := new(big.Int).SetBytes(amountData)
	decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amountFloat := new(big.Float).SetInt(value)
	amountFloat.Quo(amountFloat, new(big.Float).SetInt(decimalFactor))

	Amount := strings.TrimRight(strings.TrimRight(amountFloat.Text('f', decimals+1), "0"), ".")

	// Slack notify title and text
	var title string
	var text string

	if common.HexToAddress(vLog.Topics[0].Hex()) == common.HexToAddress("0x35d79ab81f2b2017e19afb5c5571778877782d7a8786f5907f93b0f4702f4f23") {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ETH Deposit Initialized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nAmount: %+v ETH", txHash, From, To, Amount)
	} else if common.HexToAddress(vLog.Topics[0].Hex()) == common.HexToAddress("0x2ac69ee804d9a7a0984249f508dfab7cb2534b465b6ce1580f99a38ba9c5e631") {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ETH Withdrawal Finalized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nAmount: %+v ETH", txHash, From, To, Amount)
	} else {
		title = "Unknown Event"
	}

	app.notifier.Notify(title, text)
}

func (app *App) ERC20DepAndWithEvent(vLog *types.Log) {
	log.GetLogger().Infow("Got ERC20 Deposit or Withdrawal Event", "event", vLog)

	// check the length vLog.Topics and vLog.Data
	if len(vLog.Topics) > 4 {
		log.GetLogger().Errorw("Error: Length of vLog.Topics is not as expected.")
		return
	}

	if len(vLog.Data) < 64 {
		log.GetLogger().Errorw("Error: Length of vLog.Data is not as expected.")
		return
	}

	// get symbol and decimals
	tokenAddress := vLog.Topics[1].Hex()
	tokenAddr := common.HexToAddress(tokenAddress).Hex()
	tokenInfo, found := app.tokenInfo[tokenAddr]
	if !found {
		log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddr)
		return
	}

	tokenSymbol := tokenInfo.Symbol
	tokenDecimals := tokenInfo.Decimals

	txHash := vLog.TxHash
	l1TokenAddress := common.HexToAddress(vLog.Topics[1].Hex())
	l2TokenAddress := common.HexToAddress(vLog.Topics[2].Hex())
	FromTo := common.HexToAddress(vLog.Topics[3].Hex())

	// ERC-20 deposit and withdrawal Amount
	amountData := vLog.Data[32:64]

	value := new(big.Int).SetBytes(amountData)
	decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals)), nil)
	amountFloat := new(big.Float).SetInt(value)
	amountFloat.Quo(amountFloat, new(big.Float).SetInt(decimalFactor))

	Amount := strings.TrimRight(strings.TrimRight(amountFloat.Text('f', tokenDecimals+1), "0"), ".")

	// Slack notify title and text
	var title string
	var text string

	if common.HexToAddress(vLog.Topics[0].Hex()) == common.HexToAddress("0x718594027abd4eaed59f95162563e0cc6d0e8d5b86b1c7be8b1b0ac3343d0396") {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ERC-20 Deposit Initialized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nL1TokenAddress: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2TokenAddress: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %+v%s", txHash, FromTo, FromTo, l1TokenAddress, l2TokenAddress, Amount, tokenSymbol)
	} else if common.HexToAddress(vLog.Topics[0].Hex()) == common.HexToAddress("0x3ceee06c1e37648fcbb6ed52e17b3e1f275a1f8c7b22a84b2b84732431e046b3") {
		title = fmt.Sprintf("[" + app.cfg.Network + "] [ERC-20 Withdrawal Finalized]")
		text = fmt.Sprintf("Tx: "+app.cfg.L1ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nL1TokenAddress: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2TokenAddress: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %+v%s", txHash, FromTo, FromTo, l1TokenAddress, l2TokenAddress, Amount, tokenSymbol)
	} else {
		title = "Unknown Event"
	}

	app.notifier.Notify(title, text)
}

func (app *App) L2DepAndWithEvent(vLog *types.Log) {
	log.GetLogger().Infow("Got L2 Deposit or Withdrawal Event", "event", vLog)

	// check the length vLog.Topics and vLog.Data
	if len(vLog.Topics) > 4 {
		log.GetLogger().Errorw("Error: Length of vLog.Topics is not as expected.")
		return
	}

	if len(vLog.Data) < 64 {
		log.GetLogger().Errorw("Error: Length of vLog.Data is not as expected.")
		return
	}

	// get symbol and decimals
	tokenAddress := vLog.Topics[1].Hex()
	tokenAddr := common.HexToAddress(tokenAddress).Hex()
	tokenInfo, found := app.tokenInfo[tokenAddr]
	if !found {
		log.GetLogger().Errorw("Token info not found for address", "tokenAddress", tokenAddr)
		return
	}

	tokenSymbol := tokenInfo.Symbol
	tokenDecimals := tokenInfo.Decimals

	txHash := vLog.TxHash
	l1TokenAddress := common.HexToAddress(vLog.Topics[1].Hex())
	l2TokenAddress := common.HexToAddress(vLog.Topics[2].Hex())
	FromTo := common.HexToAddress(vLog.Topics[3].Hex())

	// L2 deposit and withdrawal Amount
	amountData := vLog.Data[32:64]

	value := new(big.Int).SetBytes(amountData)
	decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals)), nil)
	amountFloat := new(big.Float).SetInt(value)
	amountFloat.Quo(amountFloat, new(big.Float).SetInt(decimalFactor))

	Amount := strings.TrimRight(strings.TrimRight(amountFloat.Text('f', int(tokenDecimals)+1), "0"), ".")

	var title string
	var text string

	if common.HexToAddress(vLog.Topics[0].Hex()) == common.HexToAddress("0xb0444523268717a02698be47d0803aa7468c00acbed2f8bd93a0459cde61dd89") {
		if common.HexToAddress(vLog.Topics[1].Hex()) == common.HexToAddress("0x0000000000000000000000000000000000000000") {
			title = fmt.Sprintf("[" + app.cfg.Network + "] [ETH Deposit Finalized]")
			text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nL1TokenAddress: Ether\nL2TokenAddress: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %+v%s", txHash, FromTo, FromTo, l2TokenAddress, Amount, tokenSymbol)
		} else {
			title = fmt.Sprintf("[" + app.cfg.Network + "] [ERC-20 Deposit Finalized]")
			text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L1ExplorerUrl+"/address/%s\nTo: "+app.cfg.L2ExplorerUrl+"/address/%s\nL1TokenAddress: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2TokenAddress: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %+v%s", txHash, FromTo, FromTo, l1TokenAddress, l2TokenAddress, Amount, tokenSymbol)
		}
	} else if common.HexToAddress(vLog.Topics[0].Hex()) == common.HexToAddress("0x73d170910aba9e6d50b102db522b1dbcd796216f5128b445aa2135272886497e") {
		if common.HexToAddress(vLog.Topics[1].Hex()) == common.HexToAddress("0x0000000000000000000000000000000000000000") {
			title = fmt.Sprintf("[" + app.cfg.Network + "] [ETH Withdrawal Initialized]")
			text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nL1TokenAddress: Ether\nL2TokenAddress: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %+v%s", txHash, FromTo, FromTo, l2TokenAddress, Amount, tokenSymbol)
		} else {
			title = fmt.Sprintf("[" + app.cfg.Network + "] [ERC-20 Withdrawal Initialized]")
			text = fmt.Sprintf("Tx: "+app.cfg.L2ExplorerUrl+"/tx/%s\nFrom: "+app.cfg.L2ExplorerUrl+"/address/%s\nTo: "+app.cfg.L1ExplorerUrl+"/address/%s\nL1TokenAddress: "+app.cfg.L1ExplorerUrl+"/token/%s\nL2TokenAddress: "+app.cfg.L2ExplorerUrl+"/token/%s\nAmount: %+v%s", txHash, FromTo, FromTo, l1TokenAddress, l2TokenAddress, Amount, tokenSymbol)
		}
	}
	app.notifier.Notify(title, text)
}

func (app *App) updateTokenInfo() error {
	data := &Data{cfg: app.cfg}
	tokenInfoMap, err := data.tokenInfoMap()
	if err != nil {
		return err
	}

	app.tokenInfo = tokenInfoMap

	return nil
}

func (app *App) Start() error {

	err := app.updateTokenInfo()
	if err != nil {
		log.GetLogger().Errorw("Failed to update token info", "err", err)
		return err
	}

	service := listener.MakeService(app.cfg.L1WsRpc)

	// L1StandardBridge ETH deposit and withdrawal
	l1BridgeETHDepositInitiated := listener.MakeEventRequest(app.cfg.L1StandardBridge, ETHDepositInitiatedEventABI, app.ETHDepAndWithEvent)
	service.AddSubscribeRequest(l1BridgeETHDepositInitiated)

	l1BridgeETHWithdrawalFinalized := listener.MakeEventRequest(app.cfg.L1StandardBridge, ETHWithdrawalFinalizedEventABI, app.ETHDepAndWithEvent)
	service.AddSubscribeRequest(l1BridgeETHWithdrawalFinalized)

	// L1StandardBridge ERC20 deposit and withdrawal
	l1BridgeERC20DepositInitiated := listener.MakeEventRequest(app.cfg.L1StandardBridge, ERC20DepositInitiatedEventABI, app.ERC20DepAndWithEvent)
	service.AddSubscribeRequest(l1BridgeERC20DepositInitiated)

	l1BridgeERC20WithdrawalFinalized := listener.MakeEventRequest(app.cfg.L1StandardBridge, ERC20WithdrawalFinalizedEventABI, app.ERC20DepAndWithEvent)
	service.AddSubscribeRequest(l1BridgeERC20WithdrawalFinalized)

	// L2StandardBridge deposit and withdrawal
	l2BridgeFinalizedDeposit := listener.MakeEventRequest(app.cfg.L2StandardBridge, DepositFinalizedEventABI, app.L2DepAndWithEvent)
	service.AddSubscribeRequest(l2BridgeFinalizedDeposit)

	l2BridgeWithdrawalRequest := listener.MakeEventRequest(app.cfg.L2StandardBridge, WithdrawalInitiatedEventABI, app.L2DepAndWithEvent)
	service.AddSubscribeRequest(l2BridgeWithdrawalRequest)

	err = service.Start()
	if err != nil {
		log.GetLogger().Errorw("Failed to start service", "err", err)
		return err
	}
	return nil
}

func New(config *Config) *App {
	slackNotifSrv := notification.MakeSlackNotificationService(config.SlackURL, 5)

	app := &App{
		cfg:       config,
		notifier:  slackNotifSrv,
		tokenInfo: make(map[string]TokenInfo),
	}

	err := app.updateTokenInfo()
	if err != nil {
		log.GetLogger().Fatalw("Failed to get token info", "error", err)
	}

	return app
}
