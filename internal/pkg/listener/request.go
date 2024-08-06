package listener

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

type Notifier interface {
	NotifyWithReTry(title string, text string)
	Notify(title string, text string) error
	Enable()
	Disable()
}
type EventRequest struct {
	contractAddress common.Address
	eventABI        string

	handler  func(vLog *types.Log) (string, string, error)
	notifier Notifier
}

func (r *EventRequest) SerializeEventRequest() string {
	hashedABI := crypto.Keccak256Hash([]byte(r.eventABI))
	return serializeEventRequestWithAddressAndABI(r.contractAddress, hashedABI)
}

func (r *EventRequest) GetRequestType() int {
	return RequestEventType
}

func (r *EventRequest) Callback(v any) {
	if v, ok := v.(*types.Log); ok {
		title, text, err := r.handler(v)
		if err != nil {
			log.GetLogger().Errorw("Failed to handle event request", "err", err, "log", v)
			return
		}

		err = r.notifier.Notify(title, text)
		if err != nil {
			log.GetLogger().Errorw("Failed to notify event request", "err", err, "log", v)
			return
		}
	}
}

func MakeEventRequest(notifier Notifier, addr string, eventABI string, handler func(vLog *types.Log) (string, string, error)) *EventRequest {
	address := common.HexToAddress(addr)
	return &EventRequest{
		contractAddress: address,
		eventABI:        eventABI,
		handler:         handler,
		notifier:        notifier,
	}
}
