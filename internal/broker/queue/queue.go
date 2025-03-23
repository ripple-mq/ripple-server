package queue

import "github.com/ripple-mq/ripple-server/pkg/utils/collection"

type Queue[T any] struct {
	q *collection.ConcurrentList[T]
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{q: collection.NewConcurrentList[T]()}
}
