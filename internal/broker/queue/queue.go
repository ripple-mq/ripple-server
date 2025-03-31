package queue

import "github.com/ripple-mq/ripple-server/pkg/utils/collection"

// Queue[T] holds thread safe queue
type Queue[T any] struct {
	q *collection.ConcurrentList[T]
}

type PayloadIF interface {
	GetID() int32
}

type Payload struct {
	Id   int32
	Data []byte
}

func (t Payload) GetID() int32 {
	return t.Id
}

type Ack struct {
	Id int32
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		q: collection.NewConcurrentList[T](),
	}
}

func (t *Queue[T]) Size() int {
	return t.q.Size()
}

func (t *Queue[T]) IsEmpty() bool {
	return t.q.Size() == 0
}

func (t *Queue[T]) Push(value T) {
	t.q.Append(value)
}

func (t *Queue[T]) Poll() T {
	return t.q.RemoveFirst()
}

func (t *Queue[T]) SubArray(start int, end int) []T {
	return t.q.Range(start, end)
}
