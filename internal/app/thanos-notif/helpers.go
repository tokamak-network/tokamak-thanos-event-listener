package thanosnotif

import (
	"errors"
	"math/big"
	"strings"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/bcclient"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/erc20"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"

	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

func formatAmount(amount *big.Int, tokenDecimals int) string {
	amountFloat := new(big.Float).SetInt(amount)
	amountFloat.Quo(amountFloat, new(big.Float).SetInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals)), nil)))
	formattedAmount := strings.TrimRight(strings.TrimRight(amountFloat.Text('f', tokenDecimals+1), "0"), ".")

	return formattedAmount
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
