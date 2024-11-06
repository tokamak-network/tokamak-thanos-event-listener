package repository

import (
	"context"
	"fmt"
	"sort"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"

	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/constant"
	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/queue"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

const (
	TwoEpochBlocks  = 64
	batchBlocksSize = uint64(10)
)

type SyncBlockMetadataKeeper interface {
	GetHead(ctx context.Context) (string, error)
	SetHead(ctx context.Context, blockHash string) error
}

type BlockChainSource interface {
	HeaderAtBlockHash(ctx context.Context, blockHash common.Hash) (*ethereumTypes.Header, error)
	GetBlocks(ctx context.Context, withLogs bool, fromBlock, toBlock uint64) ([]*types.NewBlock, error)
	GetHeader(ctx context.Context) (*ethereumTypes.Header, error)
}

type BlockKeeper struct {
	bcSource                BlockChainSource
	syncBlockMetadataKeeper SyncBlockMetadataKeeper
	head                    *ethereumTypes.Header
	q                       *queue.CircularQueue[string]
	blocks                  map[uint64]common.Hash
}

func NewBlockKeeper(ctx context.Context, bcSource BlockChainSource, syncBlockMetadataKeeper SyncBlockMetadataKeeper) (*BlockKeeper, error) {
	keeper := &BlockKeeper{
		bcSource:                bcSource,
		syncBlockMetadataKeeper: syncBlockMetadataKeeper,
		head:                    nil,
		blocks:                  make(map[uint64]common.Hash),
		q:                       queue.NewCircularQueue[string](TwoEpochBlocks),
	}

	currentBlockHash, err := syncBlockMetadataKeeper.GetHead(ctx)
	if err != nil {
		log.GetLogger().Errorw("Failed to get head", "err", err)
		return nil, err
	}

	var (
		head    *ethereumTypes.Header
		blockNo uint64
	)
	if currentBlockHash != "" {
		head, err = bcSource.HeaderAtBlockHash(ctx, common.HexToHash(currentBlockHash))
		if err != nil {
			log.GetLogger().Errorw("Failed to get head by block hash", "err", err, "hash", currentBlockHash)
			return nil, err
		}
		blockNo = head.Number.Uint64()
		keeper.head = head
	} else {
		currentHeader, err := bcSource.GetHeader(ctx)
		if err != nil {
			log.GetLogger().Errorw("Failed to get head", "err", err)
			return nil, err
		}
		currentBlockHash = currentHeader.Hash().Hex()
		blockNo = currentHeader.Number.Uint64()

		err = keeper.SetHead(ctx, currentHeader, constant.ZeroHash)
		if err != nil {
			log.GetLogger().Errorw("Failed to set head", "err", err)
			return nil, err
		}
	}

	for i := uint64(blockNo) - TwoEpochBlocks + 1; i < blockNo; i = i + batchBlocksSize {
		from := i
		to := i + batchBlocksSize - 1
		if to > blockNo-1 {
			to = blockNo - 1
		}

		blocks, err := bcSource.GetBlocks(ctx, false, from, to)
		if err != nil {
			return nil, err
		}

		for _, block := range blocks {
			keeper.enqueue(block.Header.Hash(), block.Header.Number.Uint64())
		}
	}

	keeper.enqueue(common.HexToHash(currentBlockHash), blockNo)

	log.GetLogger().Infow("Queue info", "size", keeper.q.Size(), "is_full", keeper.q.IsFull())

	return keeper, nil
}

func (bk *BlockKeeper) Head(_ context.Context) (*ethereumTypes.Header, error) {
	return bk.head, nil
}

func (bk *BlockKeeper) SetHead(ctx context.Context, header *ethereumTypes.Header, removedBlockHash common.Hash) error {
	log.GetLogger().Debug("Set head", "new", header.Hash(), "removed", removedBlockHash.Hex())
	bk.head = header

	if removedBlockHash.Cmp(constant.ZeroHash) != 0 {
		bk.q.RemoveAndEnqueue(header.Hash().String(), removedBlockHash.String())
	} else {
		bk.q.Enqueue(header.Hash().String())
	}

	err := bk.syncBlockMetadataKeeper.SetHead(ctx, header.Hash().String())
	if err != nil {
		log.GetLogger().Errorw("Failed to set head", "err", err)
		return nil
	}

	return nil
}

func (bk *BlockKeeper) Contains(header *ethereumTypes.Header) bool {
	return bk.q.Contains(header.Hash().String())
}

func (bk *BlockKeeper) GetReorgHeaders(ctx context.Context, header *ethereumTypes.Header) ([]*ethereumTypes.Header, []common.Hash, error) {
	if bk.head == nil {
		return nil, nil, nil
	}

	if header.ParentHash.Cmp(constant.ZeroHash) == 0 {
		return nil, nil, nil
	}

	if header.ParentHash.Cmp(bk.head.Hash()) == 0 {
		return nil, nil, nil
	}

	parentHash := header.ParentHash
	newHeaders := make([]*ethereumTypes.Header, 0)
	removedBlockHashes := make([]common.Hash, 0)

	for {
		if bk.q.Contains(parentHash.Hex()) {
			sort.Slice(newHeaders, func(i, j int) bool {
				blockI := newHeaders[i]
				blockJ := newHeaders[j]

				return blockI.Number.Cmp(blockJ.Number) < 0
			})

			return newHeaders, removedBlockHashes, nil
		}
		block, err := bk.bcSource.HeaderAtBlockHash(ctx, parentHash)
		if err != nil {
			log.GetLogger().Errorw("Failed to get head by block hash", "err", err)
			return nil, nil, err
		}
		if block == nil {
			return nil, nil, fmt.Errorf("block not found: %v", parentHash)
		}

		newHeaders = append(newHeaders, block)

		blockNo := block.Number.Uint64()

		if removedBlockHash, ok := bk.blocks[blockNo]; ok {
			removedBlockHashes = append(removedBlockHashes, removedBlockHash)
		} else {
			removedBlockHashes = append(removedBlockHashes, constant.ZeroHash)
		}

		parentHash = block.ParentHash
	}
}

func (bk *BlockKeeper) enqueue(blockHash common.Hash, blockNumber uint64) {
	bk.q.Enqueue(blockHash.String())
	bk.blocks[blockNumber] = blockHash
}
