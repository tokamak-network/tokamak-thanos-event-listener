package queue_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokamak-network/tokamak-thanos-event-listener/internal/pkg/queue"
)

func Test_CircularQueue(t *testing.T) {
	q := queue.NewCircularQueue[string](64)

	// Example of adding maps to the queue.
	for i := 0; i < 70; i++ {
		q.Enqueue(strconv.Itoa(i))
	}

	// Example of removing maps from the queue.
	count := 0
	for !q.IsEmpty() {
		item, err := q.Dequeue()
		if err != nil {
			t.Errorf("Failed to dequeue, err: %s", err.Error())
		} else {
			t.Logf("Item: %s", item)
		}
		count++
	}
	assert.Equal(t, 64, count)
}
