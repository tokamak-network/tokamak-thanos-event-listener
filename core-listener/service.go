package corelistener

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EventRequest struct {
	ContractAddress common.Address
	EventABI        string

	Callback func(vLog *types.Log)
}

type IEventService interface {
	RequestByKey(key string) *EventRequest
	AddEventFilter(eventRequest *EventRequest) bool
	StartListen() error
	Call(msg ethereum.CallMsg) ([]byte, error)
	GetBlockByHash(blockHash common.Hash) (uint64, error)
}
