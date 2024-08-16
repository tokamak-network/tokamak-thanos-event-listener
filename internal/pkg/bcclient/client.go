package bcclient

import (
	"context"
	"math/big"
	"net/http"
	"time"

	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"
	"golang.org/x/sync/errgroup"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

type Client struct {
	defaultClient *ethclient.Client
	wsClient      *ethclient.Client
	chainID       *big.Int
}

func New(ctx context.Context, wsURL, rpcURL string) (*Client, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	ethClient, err := initEthClient(ctx, rpcURL, httpClient)
	if err != nil {
		return nil, err
	}

	wsClient, err := initEthClient(ctx, wsURL, httpClient)
	if err != nil {
		return nil, err
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		defaultClient: ethClient,
		wsClient:      wsClient,
		chainID:       chainID,
	}, nil
}

func initEthClient(ctx context.Context, url string, httpClient *http.Client) (*ethclient.Client, error) {
	rpcClient, err := rpc.DialOptions(ctx, url, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	ethClient := ethclient.NewClient(rpcClient)
	return ethClient, nil
}

func (c *Client) GetClient() *ethclient.Client {
	return c.defaultClient
}

func (c *Client) SubscribeNewHead(ctx context.Context, newHeadCh chan<- *ethereumTypes.Header) (ethereum.Subscription, error) {
	return c.wsClient.SubscribeNewHead(ctx, newHeadCh)
}

func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	return c.defaultClient.BlockNumber(ctx)
}

func (c *Client) GetHeader(ctx context.Context) (*ethereumTypes.Header, error) {
	return c.defaultClient.HeaderByNumber(ctx, nil)
}

func (c *Client) HeaderAtBlockNumber(ctx context.Context, blockNo uint64) (*ethereumTypes.Header, error) {
	headerAtBlockNo, err := c.defaultClient.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNo))
	if err != nil {
		return nil, err
	}

	return headerAtBlockNo, nil
}

func (c *Client) GetLogs(ctx context.Context, blockHash common.Hash) ([]ethereumTypes.Log, error) {
	var err error
	var logs []ethereumTypes.Log
	for i := 0; i < 3; i++ {
		query := ethereum.FilterQuery{
			BlockHash: &blockHash,
		}

		// Get the logs
		logs, err = c.defaultClient.FilterLogs(ctx, query)
		if err != nil {
			log.GetLogger().Errorw("Failed to retrieve logs", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		return logs, nil
	}

	return nil, err
}

func (c *Client) HeaderAtBlockHash(ctx context.Context, blockHash common.Hash) (*ethereumTypes.Header, error) {
	headerAtBlockHash, err := c.defaultClient.HeaderByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}

	return headerAtBlockHash, nil
}

func (c *Client) GetBlocks(ctx context.Context, withLogs bool, fromBlock, toBlock uint64) ([]*types.NewBlock, error) {
	log.GetLogger().Infow("Fetch blocks info", "from_block", fromBlock, "to_block", toBlock)
	totalBlocks := toBlock - fromBlock + 1

	blocks := make([]*types.NewBlock, totalBlocks)

	g, _ := errgroup.WithContext(ctx)
	for index := uint64(0); index < totalBlocks; index++ {
		index := index
		blockNo := index + fromBlock

		g.Go(func() error {
			header, err := c.HeaderAtBlockNumber(ctx, blockNo)
			if err != nil {
				log.GetLogger().Errorw("Failed to get block header", "err", err)
				return err
			}

			blocks[index] = &types.NewBlock{
				Header: header,
			}

			if withLogs {
				logs, err := c.GetLogs(ctx, header.Hash())
				if err != nil {
					log.GetLogger().Errorw("Failed to get block logs", "err", err)
					return err
				}
				blocks[index].Logs = logs
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		log.GetLogger().Errorw("Failed to get the block header", "err", err)

		return nil, err
	}

	if len(blocks) == 0 {
		return nil, nil
	}

	return blocks, nil
}
