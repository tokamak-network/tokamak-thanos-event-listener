package erc20

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

func FetchTokenInfo(client *ethclient.Client, tokenAddress string) (*types.Token, error) {
	erc20Instance, err := NewErc20(common.HexToAddress(tokenAddress), client)
	if err != nil {
		log.GetLogger().Errorw("Failed to create erc20 instance", "error", err)
		return nil, err
	}

	symbol, err := erc20Instance.Symbol(nil)
	if err != nil {
		log.GetLogger().Errorw("Failed to get symbol", "error", err)
		return nil, err
	}

	decimals, err := erc20Instance.Decimals(nil)
	if err != nil {
		log.GetLogger().Errorw("Failed to get decimals", "error", err)
		return nil, err
	}

	return &types.Token{
		Decimals: int(decimals),
		Symbol:   symbol,
		Address:  tokenAddress,
	}, nil
}
