package queue

import "github.com/ripple-mq/ripple-server/pkg/utils/collection"

type Queue[T any] struct {
	q *collection.ConcurrentList[T]
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		q: collection.NewConcurrentList[T](),
	}
}

func (t *Queue[T]) Push(value T) {
	t.q.Append(value)
}

func (t *Queue[T]) SubArray(start int, end int) []T {
	return t.q.Range(start, end)
}
