package listener

import (
	"github.com/bits-and-blooms/bloom/v3"
)

type CounterBloom struct {
	counter int
	max     int
	bloom   *bloom.BloomFilter
}

func (filter *CounterBloom) Add(data []byte) {
	filter.counter++
	if filter.counter > filter.max {
		filter.counter = 1
		filter.bloom.ClearAll()
	}
	filter.bloom.Add(data)
}

func (filter *CounterBloom) Test(data []byte) bool {
	return filter.bloom.Test(data)
}

func MakeDefaultCounterBloom() *CounterBloom {
	maxItem := 60000
	return MakeCounterBloom(maxItem)
}

func MakeCounterBloom(maxItem int) *CounterBloom {
	return &CounterBloom{counter: 0, max: maxItem, bloom: bloom.NewWithEstimates(uint(maxItem), 0.00000001)}
}
