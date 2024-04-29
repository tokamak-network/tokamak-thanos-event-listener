package listener

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

var (
	RequestEventType = 1
)

type RequestSubscriber interface {
	GetRequestType() int
	SerializeEventRequest() string
	Callback(item interface{})
}

type EventService struct {
	host       string
	client     *ethclient.Client
	requests   []RequestSubscriber
	requestMap map[string]RequestSubscriber
	startBlock *big.Int
	filter     *CounterBloom
}

func MakeService(host string) *EventService {
	service := &EventService{host: host, filter: MakeDefaultCounterBloom()}
	service.initialize()
	return service
}

func MakeServiceWithStartBlock(host string, start *big.Int) *EventService {
	service := &EventService{host: host, startBlock: start, filter: MakeDefaultCounterBloom()}
	service.initialize()
	return service
}

func (service *EventService) initialize() {
	service.requests = make([]RequestSubscriber, 0)
	service.requestMap = make(map[string]RequestSubscriber)
}

func (service *EventService) existRequest(request RequestSubscriber) bool {
	key := request.SerializeEventRequest()
	_, ok := service.requestMap[key]
	return ok
}

func (service *EventService) RequestByKey(key string) RequestSubscriber {
	request, ok := service.requestMap[key]
	if ok {
		return request
	} else {
		return nil
	}
}

func (service *EventService) AddSubscribeRequest(request RequestSubscriber) {
	if service.existRequest(request) {
		return
	}
	service.requests = append(service.requests, request)
	key := request.SerializeEventRequest()
	service.requestMap[key] = request
}

func (service *EventService) CanProcess(log *types.Log) bool {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(log)
	if err != nil {
		return false
	}
	data := buf.Bytes()

	if service.filter.Test(data) {
		return false
	}
	service.filter.Add(data)
	return true
}

func (service *EventService) Start() error {
	log.GetLogger().Infow("Start to listen", "host", service.host)
	client, err := ethclient.Dial(service.host)
	if err != nil {
		return err
	}

	log.GetLogger().Infow("Connected to", "host", service.host)

	service.client = client

	fromBlock := service.startBlock

	addresses := CalculateAddresses(service.requests)

	log.GetLogger().Infow("Listen to these addresses", "addresses", addresses, "from_block", fromBlock)

	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		Addresses: addresses,
	}

	logsCh := make(chan types.Log)
	defer close(logsCh)
	sub, err := service.client.SubscribeFilterLogs(context.Background(), query, logsCh)
	if err != nil {
		log.GetLogger().Errorw("Failed to subscribe filter logs", "err", err)
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			log.GetLogger().Errorw("Failed to listen the subscription", "err", err)
			return err
		case vLog := <-logsCh:
			key := serializeEventRequestWithAddressAndABI(vLog.Address, vLog.Topics[0])
			request := service.RequestByKey(key)
			if request != nil {
				if service.CanProcess(&vLog) {
					request.Callback(&vLog)
				}
			}
		}
	}
}

func (service *EventService) Call(msg ethereum.CallMsg) ([]byte, error) {
	var data []byte
	var err error
	for i := 0; i < 5; i++ {
		data, err = service.client.CallContract(context.Background(), msg, nil)
		if err != nil {
			continue
		}
		log.GetLogger().Infow("Result when calling the contract", "data", string(data))
		break
	}
	return data, err
}

func (service *EventService) GetBlockByHash(blockHash common.Hash) (*types.Block, error) {
	var block *types.Block
	var err error
	for i := 0; i < 5; i++ {
		block, err = service.client.BlockByHash(context.Background(), blockHash)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.GetLogger().Errorw("Failed to retrieve block", "err", err)
		return nil, err
	}
	return block, nil
}

func serializeEventRequestWithAddressAndABI(address common.Address, hashedABI common.Hash) string {
	result := fmt.Sprintf("%s:%s", address.String(), hashedABI)
	return result
}
