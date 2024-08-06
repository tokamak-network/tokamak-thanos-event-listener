package bcclient

import (
	"context"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

type Client struct {
	defaultClient *ethclient.Client
	chainID       *big.Int
}

func New(ctx context.Context, rpcURL string) (*Client, error) {
	httpClient := &http.Client{
		Timeout: 3 * time.Second,
	}
	rpcClient, err := rpc.DialOptions(ctx, rpcURL, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	ethClient := ethclient.NewClient(rpcClient)

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		defaultClient: ethClient,
		chainID:       chainID,
	}, nil
}

func (c *Client) GetClient() *ethclient.Client {
	return c.defaultClient
}

func (c *Client) SubscribeNewHead(ctx context.Context, newHeadCh chan<- *types.Header) (ethereum.Subscription, error) {
	return c.defaultClient.SubscribeNewHead(ctx, newHeadCh)
}

func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	return c.defaultClient.BlockNumber(ctx)
}

func (c *Client) HeaderAtBlockNumber(ctx context.Context, blockNo uint64) (*types.Header, error) {
	headerAtBlockNo, err := c.defaultClient.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNo))
	if err != nil {
		return nil, err
	}

	return headerAtBlockNo, nil
}

func (c *Client) GetLogs(ctx context.Context, blockHash common.Hash) ([]types.Log, error) {
	timeOutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := ethereum.FilterQuery{
		BlockHash: &blockHash,
	}

	// Get the logs
	logs, err := c.defaultClient.FilterLogs(timeOutCtx, query)
	if err != nil {
		log.GetLogger().Errorw("Failed to retrieve logs", "blockHash", blockHash.Hex(), "err", err)
		return nil, err
	}

	return logs, nil
}

func (c *Client) HeaderAtBlockHash(ctx context.Context, blockHash common.Hash) (*types.Header, error) {
	headerAtBlockHash, err := c.defaultClient.HeaderByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}

	return headerAtBlockHash, nil
}
