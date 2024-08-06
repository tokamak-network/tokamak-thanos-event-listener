package repository

import (
	"context"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/constant"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/queue"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

const (
	TwoEpochBlocks = 64
)

type SyncBlockMetadataKeeper interface {
	GetHead(ctx context.Context) (string, error)
	SetHead(ctx context.Context, blockHash string) error
}

type BlockChainSource interface {
	HeaderAtBlockHash(ctx context.Context, blockHash common.Hash) (*types.Header, error)
}

type BlockKeeper struct {
	bcSource                BlockChainSource
	syncBlockMetadataKeeper SyncBlockMetadataKeeper
	head                    *types.Header
	q                       *queue.CircularQueue[string]
}

func NewBlockKeeper(ctx context.Context, bcSource BlockChainSource, syncBlockMetadataKeeper SyncBlockMetadataKeeper) (*BlockKeeper, error) {
	blockHash, err := syncBlockMetadataKeeper.GetHead(ctx)
	if err != nil {
		log.GetLogger().Errorw("Failed to get head", "err", err)
		return nil, err
	}

	circularQueue := queue.NewCircularQueue[string](TwoEpochBlocks)

	var (
		head *types.Header
	)
	if blockHash != "" {
		head, err = bcSource.HeaderAtBlockHash(ctx, common.HexToHash(blockHash))
		if err != nil {
			return nil, err
		}
		circularQueue.Enqueue(head.Hash().String())
	}
	keeper := &BlockKeeper{
		q:                       circularQueue,
		bcSource:                bcSource,
		head:                    head,
		syncBlockMetadataKeeper: syncBlockMetadataKeeper,
	}

	return keeper, nil
}

func (bk *BlockKeeper) Head(_ context.Context) (*types.Header, error) {
	return bk.head, nil
}

func (bk *BlockKeeper) SetHead(ctx context.Context, header *types.Header) error {
	bk.head = header
	bk.q.Enqueue(header.Hash().String())

	err := bk.syncBlockMetadataKeeper.SetHead(ctx, header.Hash().String())
	if err != nil {
		log.GetLogger().Errorw("Failed to set head", "err", err)
		return nil
	}
	return nil
}

func (bk *BlockKeeper) Contains(header *types.Header) bool {
	return bk.q.Contains(header.Hash().String())
}

func (bk *BlockKeeper) GetReorgHeaders(ctx context.Context, header *types.Header) ([]*types.Header, error) {
	if bk.head == nil {
		return nil, nil
	}

	if header.ParentHash.Cmp(constant.ZeroHash) == 0 {
		return nil, nil
	}

	if header.ParentHash.Cmp(bk.head.Hash()) == 0 {
		return nil, nil
	}

	parentHash := header.ParentHash
	reorgHeaders := make([]*types.Header, 0)

	for {
		block, err := bk.bcSource.HeaderAtBlockHash(ctx, parentHash)
		if err != nil {
			return nil, err
		}

		if bk.q.Contains(block.Hash().String()) {
			sort.Slice(reorgHeaders, func(i, j int) bool {
				blockI := reorgHeaders[i]
				blockJ := reorgHeaders[j]

				return blockI.Number.Cmp(blockJ.Number) < 0
			})

			return reorgHeaders, nil
		}
		reorgHeaders = append(reorgHeaders, block)

		parentHash = block.ParentHash
	}
}
