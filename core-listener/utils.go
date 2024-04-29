package corelistener

import "github.com/ethereum/go-ethereum/common"

func CalculateAddresses(requests []SubcribeRequest) []common.Address {
	encountered := map[common.Address]bool{}
	result := []common.Address{}

	for _, v := range requests {
		if v.GetRequestType() == REQUEST_EVENT_TYPE {
			eventRequest, _ := v.(*EventRequest)
			if !encountered[eventRequest.contractAddress] {
				encountered[eventRequest.contractAddress] = true
				result = append(result, eventRequest.contractAddress)
			}
		}
	}
	return result
}
