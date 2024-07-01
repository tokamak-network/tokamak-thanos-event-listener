package thanosnotif

import (
	"math/big"
	"strings"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

func (app *App) formatAmount(amount *big.Int, tokenDecimals int) string {
	amountFloat := new(big.Float).SetInt(amount)
	amountFloat.Quo(amountFloat, new(big.Float).SetInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals)), nil)))
	formattedAmount := strings.TrimRight(strings.TrimRight(amountFloat.Text('f', tokenDecimals+1), "0"), ".")

	return formattedAmount
}

func (app *App) getBridgeFilterers() (l1BridgeFilterer *bindings.L1StandardBridgeFilterer, l2BridgeFilterer *bindings.L2StandardBridgeFilterer, err error) {
	client, err := ethclient.Dial(app.cfg.L1Rpc)
	if err != nil {
		log.GetLogger().Errorw("Failed to connect to Ethereum client", "error", err)
		return nil, nil, err
	}

	l1BridgeFilterer, err = bindings.NewL1StandardBridgeFilterer(common.HexToAddress(app.cfg.L1StandardBridge), client)
	if err != nil {
		log.GetLogger().Errorw("L1StandardBridgeFilterer instance fail", "error", err)
		return nil, nil, err
	}

	l2BridgeFilterer, err = bindings.NewL2StandardBridgeFilterer(common.HexToAddress(app.cfg.L2StandardBridge), client)
	if err != nil {
		log.GetLogger().Errorw("L2StandardBridgeFilterer instance fail", "error", err)
		return nil, nil, err
	}

	return l1BridgeFilterer, l2BridgeFilterer, nil
}
