package listener

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/types"
)

var (
	RequestEventType = 1
)

const (
	MaxBatchBlocksSize = 10
)

type RequestSubscriber interface {
	GetRequestType() int
	SerializeEventRequest() string
	Callback(item any)
}

type BlockKeeper interface {
	Head(ctx context.Context) (*ethereumTypes.Header, error)
	SetHead(ctx context.Context, header *ethereumTypes.Header) error
	Contains(header *ethereumTypes.Header) bool
	GetReorgHeaders(ctx context.Context, header *ethereumTypes.Header) ([]*ethereumTypes.Header, error)
}

type BlockChainSource interface {
	SubscribeNewHead(ctx context.Context, newHeadCh chan<- *ethereumTypes.Header) (ethereum.Subscription, error)
	BlockNumber(ctx context.Context) (uint64, error)
	HeaderAtBlockNumber(ctx context.Context, blockNo uint64) (*ethereumTypes.Header, error)
	GetLogs(ctx context.Context, blockHash common.Hash) ([]ethereumTypes.Log, error)
}

type EventService struct {
	l           *zap.SugaredLogger
	bcClient    BlockChainSource
	blockKeeper BlockKeeper
	requestMap  map[string]RequestSubscriber
	filter      *CounterBloom
	sub         ethereum.Subscription
}

func MakeService(name string, bcClient BlockChainSource, keeper BlockKeeper) (*EventService, error) {
	service := &EventService{
		l:           log.GetLogger().Named(name),
		bcClient:    bcClient,
		blockKeeper: keeper,
		filter:      MakeDefaultCounterBloom(),
		requestMap:  make(map[string]RequestSubscriber),
	}

	return service, nil
}

func (s *EventService) existRequest(request RequestSubscriber) bool {
	key := request.SerializeEventRequest()
	_, ok := s.requestMap[key]
	return ok
}

func (s *EventService) RequestByKey(key string) RequestSubscriber {
	request, ok := s.requestMap[key]
	if ok {
		return request
	}
	return nil
}

func (s *EventService) AddSubscribeRequest(request RequestSubscriber) {
	if s.existRequest(request) {
		return
	}
	key := request.SerializeEventRequest()
	s.requestMap[key] = request
}

func (s *EventService) CanProcess(log *ethereumTypes.Log) bool {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(log)
	if err != nil {
		return false
	}
	data := buf.Bytes()

	if s.filter.Test(data) {
		return false
	}
	s.filter.Add(data)
	return true
}

func (s *EventService) Start(ctx context.Context) error {
	oldBlocksCh := make(chan *types.NewBlock)

	errCh := make(chan error, 1)
	s.l.Infow("Start to sync old blocks")

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := s.syncOldBlocks(ctx, oldBlocksCh)
		defer close(oldBlocksCh)

		if err != nil {
			s.l.Errorw("Failed to sync old blocks", "err", err)
			return err
		}

		return nil
	})

	for oldBlock := range oldBlocksCh {
		err := s.handleNewBlock(ctx, oldBlock)
		if err != nil {
			s.l.Errorw("Failed to handle the old block", "err", err)
			return err
		}
	}

	if err := g.Wait(); err != nil {
		s.l.Errorw("Failed to sync old blocks", "err", err)
		return err
	}

	s.sub = event.ResubscribeErr(10, func(ctx context.Context, err error) (event.Subscription, error) {
		if err != nil {
			s.l.Errorw("Failed to re-subscribe the event", "err", err)
		}

		return s.subscribeNewHead(ctx)
	})

	go func() {
		err, ok := <-s.sub.Err()
		if !ok {
			return
		}
		s.l.Errorw("Failed to subscribe new head", "err", err)

		errCh <- err
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			s.l.Errorw("Failed to re-subscribe the event", "err", err)
			return err
		}
	}
}

func (s *EventService) subscribeNewHead(
	_ context.Context,
) (ethereum.Subscription, error) {
	headChanges := make(chan *ethereumTypes.Header, 64)

	ctx := context.Background()

	sub, err := s.bcClient.SubscribeNewHead(ctx, headChanges)
	if err != nil {
		return nil, err
	}
	s.l.Infow("Start process new head")

	newSub := event.NewSubscription(func(quit <-chan struct{}) error {
		eventsCtx, cancelFunc := context.WithCancel(ctx)
		defer sub.Unsubscribe()
		defer cancelFunc()

		go func() {
			select {
			case <-quit:
				cancelFunc()
			case <-eventsCtx.Done(): // don't wait for quit signal if we closed for other reasons.
				return
			}
		}()

		for {
			select {
			case newHead := <-headChanges:
				s.l.Infow("New head received", "header", newHead.Number)

				logs, err := s.bcClient.GetLogs(ctx, newHead.Hash())
				if err != nil {
					s.l.Errorw("Failed to filter logs", "err", err)
					return err
				}

				err = s.handleNewBlock(ctx, &types.NewBlock{
					Logs:   logs,
					Header: newHead,
				})

				if err != nil {
					s.l.Errorw("Failed to handle the new head", "err", err)
					return err
				}

			case <-eventsCtx.Done():
				return nil
			case subErr := <-sub.Err():
				return subErr
			}
		}
	})

	return newSub, nil
}

func (s *EventService) handleNewBlock(ctx context.Context, newBlock *types.NewBlock) error {
	if newBlock == nil {
		return nil
	}

	newHeader := newBlock.Header
	reorgedBlocks, err := s.handleReorgBlocks(ctx, newHeader)
	if err != nil {
		s.l.Errorw("Failed to handle re-org blocks", "err", err)
		return err
	}

	blocks := make([]*types.NewBlock, 0)
	if len(reorgedBlocks) > 0 {
		blocks = append(blocks, reorgedBlocks...)
	}

	blocks = append(blocks, newBlock)

	for _, block := range blocks {
		err = s.filterEventsAndNotify(ctx, block.Logs)
		if err != nil {
			s.l.Errorw("Failed to handle block", "err", err, "block", block)
			return err
		}
		err = s.blockKeeper.SetHead(ctx, block.Header)
		if err != nil {
			s.l.Errorw("Failed to set head on the keeper", "err", err, "block", block)
			return err
		}
	}

	return nil
}

func (s *EventService) filterEventsAndNotify(_ context.Context, logs []ethereumTypes.Log) error {
	for _, l := range logs {
		if len(l.Topics) == 0 {
			continue
		}

		key := serializeEventRequestWithAddressAndABI(l.Address, l.Topics[0])
		request := s.RequestByKey(key)

		if request == nil {
			continue
		}

		if !s.CanProcess(&l) {
			continue
		}

		request.Callback(&l)
	}
	return nil
}

func (s *EventService) syncOldBlocks(ctx context.Context, headCh chan *types.NewBlock) error {
	onchainBlockNo, err := s.bcClient.BlockNumber(ctx)
	if err != nil {
		return err
	}

	consumingBlock, err := s.blockKeeper.Head(ctx)
	if err != nil {
		s.l.Errorw("Failed to get block head from keeper", "err", err)
		return err
	}

	if consumingBlock == nil {
		return nil
	}

	consumedBlockNo := consumingBlock.Number.Uint64()

	if consumedBlockNo > onchainBlockNo {
		return nil
	}

	s.l.Infow("Fetch old blocks", "consumed_block", consumedBlockNo, "onchain_block", onchainBlockNo)

	blocksNeedToConsume := onchainBlockNo - consumedBlockNo

	totalBatches := int(math.Ceil(float64(blocksNeedToConsume) / float64(MaxBatchBlocksSize)))

	s.l.Infow("Total batches", "total", totalBatches)
	skip := consumedBlockNo + 1
	for i := 0; i < totalBatches; i++ {
		fromBlock := skip
		toBlock := skip + MaxBatchBlocksSize - 1

		if toBlock > onchainBlockNo {
			toBlock = onchainBlockNo
		}

		blocks, err := s.getBlocks(ctx, fromBlock, toBlock)
		if err != nil {
			return err
		}

		for _, oldHead := range blocks {
			headCh <- oldHead
		}
		skip = toBlock + 1
	}

	return nil
}

func (s *EventService) getBlocks(ctx context.Context, fromBlock, toBlock uint64) ([]*types.NewBlock, error) {
	s.l.Infow("Fetch blocks info", "from_block", fromBlock, "to_block", toBlock)
	totalBlocks := toBlock - fromBlock + 1

	blocks := make([]*types.NewBlock, totalBlocks)

	g, _ := errgroup.WithContext(ctx)
	for index := uint64(0); index < totalBlocks; index++ {
		index := index
		blockNo := index + fromBlock

		g.Go(func() error {
			header, err := s.bcClient.HeaderAtBlockNumber(ctx, blockNo)
			if err != nil {
				s.l.Errorw("Failed to get block header", "err", err)
				return err
			}

			logs, err := s.bcClient.GetLogs(ctx, header.Hash())
			if err != nil {
				s.l.Errorw("Failed to get block logs", "err", err)
				return err
			}

			blocks[index] = &types.NewBlock{
				Logs:   logs,
				Header: header,
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		s.l.Errorw("Failed to get the block header", "err", err)

		return nil, err
	}

	if len(blocks) == 0 {
		return nil, nil
	}

	return blocks, nil
}

func (s *EventService) handleReorgBlocks(ctx context.Context, newHeader *ethereumTypes.Header) ([]*types.NewBlock, error) {
	reorgedHeaders, err := s.blockKeeper.GetReorgHeaders(ctx, newHeader)
	if err != nil {
		s.l.Errorw("Failed to handle reorg blocks", "err", err)
		return nil, err
	}

	if len(reorgedHeaders) == 0 {
		return nil, nil
	}
	var g errgroup.Group

	reorgedBlocks := make([]*types.NewBlock, len(reorgedHeaders))
	for i, reorgedHeader := range reorgedHeaders {
		s.l.Infow("Detect reorg block", "block", reorgedHeader.Number.Uint64())
		i := i
		reorgedHeader := reorgedHeader

		g.Go(func() error {
			blockHash := reorgedHeader.Hash()
			reorgedLogs, errLogs := s.bcClient.GetLogs(ctx, blockHash)
			if errLogs != nil {
				s.l.Errorw("Failed to get logs", "err", errLogs)
				return errLogs
			}

			reorgedBlocks[i] = &types.NewBlock{
				Header: reorgedHeader,
				Logs:   reorgedLogs,
			}

			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return nil, err
	}

	return reorgedBlocks, nil
}

func serializeEventRequestWithAddressAndABI(address common.Address, hashedABI common.Hash) string {
	result := fmt.Sprintf("%s:%s", address.String(), hashedABI)
	return result
}
