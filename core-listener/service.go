package corelistener

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var REQUEST_EVENT_TYPE = 1

type SubcribeRequest interface {
	GetRequestType() int
	SerializeEventRequest() string
	Callback(item interface{})
}

type EventRequest struct {
	contractAddress common.Address
	eventABI        string

	handler func(vLog *types.Log)
}

func SerializeEventRequestWithAddressAndABI(address common.Address, hashedABI common.Hash) string {
	result := fmt.Sprintf("%s:%s", address.String(), hashedABI)
	return result
}

func (request *EventRequest) SerializeEventRequest() string {
	hashedABI := crypto.Keccak256Hash([]byte(request.eventABI))
	return SerializeEventRequestWithAddressAndABI(request.contractAddress, hashedABI)
}

func (request *EventRequest) GetRequestType() int {
	return REQUEST_EVENT_TYPE
}

func (request *EventRequest) Callback(v interface{}) {
	if v, ok := v.(*types.Log); ok {
		request.handler(v)
	}
}

func MakeEventRequest(addr string, eventABI string, handler func(vLog *types.Log)) *EventRequest {
	address := common.HexToAddress(addr)
	return &EventRequest{contractAddress: address, eventABI: eventABI, handler: handler}
}

type IEventService interface {
	RequestByKey(key string) SubcribeRequest
	AddSubscribeRequest(request SubcribeRequest) bool
	StartListen() error
	Call(msg ethereum.CallMsg) ([]byte, error)
	GetBlockByHash(blockHash common.Hash) (*types.Block, error)
}

type EventService struct {
	host       string
	client     *ethclient.Client
	sub        ethereum.Subscription
	logs       chan types.Log
	requests   []SubcribeRequest
	requestMap map[string]SubcribeRequest
	startBlock *big.Int
	filter     *CounterBloom
}

func (service *EventService) initialize() {
	service.requests = make([]SubcribeRequest, 0)
	service.requestMap = make(map[string]SubcribeRequest)
}

func (service *EventService) existRequest(request SubcribeRequest) bool {
	key := request.SerializeEventRequest()
	_, ok := service.requestMap[key]
	return ok
}

func (service *EventService) RequestByKey(key string) SubcribeRequest {
	request, ok := service.requestMap[key]
	if ok {
		return request
	} else {
		return nil
	}
}

func (service *EventService) AddSubscribeRequest(request SubcribeRequest) bool {
	if !service.existRequest(request) {
		service.requests = append(service.requests, request)
		key := request.SerializeEventRequest()
		service.requestMap[key] = request
		return true
	}
	return false
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
	fmt.Println("Started Listen to ", service.host)
	client, err := ethclient.Dial(service.host)
	fmt.Println("client:", client)
	fmt.Println("err:", err)
	if err == nil {
		fmt.Println("Connected to Host: ", service.host)
		service.client = client
	} else {
		fmt.Println("Failed Connected to Host: ", service.host, "| Err:", err)
		return err
	}

	fromBlock := service.startBlock

	addresses := CalculateAddresses(service.requests)

	fmt.Println("Address list for listen: ", addresses)
	fmt.Println("FromBlock: ", fromBlock)

	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		Addresses: addresses,
	}

	service.logs = make(chan types.Log)
	service.sub, err = service.client.SubscribeFilterLogs(context.Background(), query, service.logs)
	if err != nil {
		fmt.Println("SubscribeFilterLogs err: ", err)
		return err
	}

	for {
		select {
		case err := <-service.sub.Err():
			fmt.Println("Err when listening:", err)
			return err
		case vLog := <-service.logs:
			key := SerializeEventRequestWithAddressAndABI(vLog.Address, vLog.Topics[0])
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
		if err == nil {
			fmt.Println(data)
			break
		}
	}
	return data, err
}

func (service *EventService) GetBlockByHash(blockHash common.Hash) (*types.Block, error) {
	client := service.client
	var block *types.Block
	var err error
	for i := 0; i < 5; i++ {
		block, err = client.BlockByHash(context.Background(), blockHash)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Println("Failed to retrieve block:", err)
		return nil, err
	}
	return block, nil
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
