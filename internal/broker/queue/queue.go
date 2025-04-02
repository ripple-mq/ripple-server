package queue

import "github.com/ripple-mq/ripple-server/pkg/utils/collection"

// Queue[T] represents a thread-safe queue that stores elements of any type.
type Queue[T any] struct {
	q *collection.ConcurrentList[T]
}

// PayloadIF defines an interface for types that provide a method to retrieve an ID.
type PayloadIF interface {
	GetID() int32
}

// Payload represents a data structure holding an ID and data.
type Payload struct {
	Id   int32  // Unique identifier for the payload
	Data []byte // Actual data of the payload
}

// GetID returns the ID of the payload.
func (t Payload) GetID() int32 {
	return t.Id
}

// Ack represents an acknowledgment containing an ID.
type Ack struct {
	Id int32
}

// NewQueue[T] creates a new, empty Queue of the given type T.
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		q: collection.NewConcurrentList[T](),
	}
}

// Size returns the number of elements in the queue.
func (t *Queue[T]) Size() int {
	return t.q.Size()
}

// IsEmpty checks if the queue is empty.
func (t *Queue[T]) IsEmpty() bool {
	return t.q.Size() == 0
}

// Push adds a new element to the end of the queue.
func (t *Queue[T]) Push(value T) {
	t.q.Append(value)
}

// Poll removes and returns the first element from the queue.
func (t *Queue[T]) Poll() T {
	return t.q.RemoveFirst()
}

// SubArray returns a slice of elements from the queue, between the given indices.
func (t *Queue[T]) SubArray(start int, end int) []T {
	return t.q.Range(start, end)
}
