package listener

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type EventRequest struct {
	contractAddress common.Address
	eventABI        string

	handler func(vLog *types.Log)
}

func (request *EventRequest) SerializeEventRequest() string {
	hashedABI := crypto.Keccak256Hash([]byte(request.eventABI))
	return serializeEventRequestWithAddressAndABI(request.contractAddress, hashedABI)
}

func (request *EventRequest) GetRequestType() int {
	return RequestEventType
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
