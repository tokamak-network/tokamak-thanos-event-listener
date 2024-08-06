package queue

import (
	"errors"
)

// CircularQueue represents a circular queue with a fixed size.
type CircularQueue[T comparable] struct {
	size       int
	data       []T
	uniqueData map[T]bool
	front      int
	rear       int
	count      int
}

// NewCircularQueue creates a new CircularQueue with the specified size.
func NewCircularQueue[T comparable](size int) *CircularQueue[T] {
	return &CircularQueue[T]{
		size:       size,
		data:       make([]T, size),
		uniqueData: make(map[T]bool),
		rear:       -1, // rear starts at -1 to handle the first increment correctly
	}
}

// Enqueue adds an element to the end of the queue.
// If the queue is full, it overwrites the oldest element.
func (q *CircularQueue[T]) Enqueue(value T) {
	if q.IsFull() {
		// When the queue is full, we overwrite the oldest item.
		q.front = (q.front + 1) % q.size
	} else {
		q.count++
	}
	q.rear = (q.rear + 1) % q.size
	q.data[q.rear] = value
	q.uniqueData[value] = true
}

// Dequeue removes and returns the element at the front of the queue.
func (q *CircularQueue[T]) Dequeue() (T, error) {
	var zeroValue T
	if q.IsEmpty() {
		return zeroValue, errors.New("queue is empty")
	}
	value := q.data[q.front]
	q.data[q.front] = zeroValue // Clear the slot
	delete(q.uniqueData, value)
	q.front = (q.front + 1) % q.size
	q.count--

	return value, nil
}

// GetHead returns the element at the front of the queue without removing it.
func (q *CircularQueue[T]) GetHead() (T, error) {
	var zeroValue T
	if q.IsEmpty() {
		return zeroValue, errors.New("queue is empty")
	}

	return q.data[q.front], nil
}

// Contains checks if a specified element exists in the queue.
func (q *CircularQueue[T]) Contains(value T) bool {
	return q.uniqueData[value]
}

// Size returns the number of elements in the queue.
func (q *CircularQueue[T]) Size() int {
	return q.count
}

// IsEmpty checks if the queue is empty.
func (q *CircularQueue[T]) IsEmpty() bool {
	return q.count == 0
}

// IsFull checks if the queue is full.
func (q *CircularQueue[T]) IsFull() bool {
	return q.count == q.size
}
