package listener

import (
	"github.com/ethereum/go-ethereum/common"
)

func CalculateAddresses(requests []RequestSubscriber) []common.Address {
	encountered := map[common.Address]bool{}
	result := make([]common.Address, 0)

	for _, v := range requests {
		if v.GetRequestType() == RequestEventType {
			eventRequest, _ := v.(*EventRequest)
			if !encountered[eventRequest.contractAddress] {
				encountered[eventRequest.contractAddress] = true
				result = append(result, eventRequest.contractAddress)
			}
		}
	}
	return result
}
