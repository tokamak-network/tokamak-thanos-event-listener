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

func Test_CircularQueue_Remove(t *testing.T) {
	q := queue.NewCircularQueue[string](64)

	for i := 0; i < 70; i++ {
		q.Enqueue(strconv.Itoa(i))
	}

	q.Remove("68")
	assert.Equal(t, false, q.Contains("68"))
	assert.Equal(t, 63, q.Size())

	q.Remove("69")
	assert.Equal(t, false, q.Contains("69"))
	assert.Equal(t, 62, q.Size())

	q.RemoveAndEnqueue("70", "67")
	assert.Equal(t, true, q.Contains("70"))
	assert.Equal(t, false, q.Contains("67"))
	assert.Equal(t, 62, q.Size())
}
