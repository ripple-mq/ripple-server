package collection

import (
	"sync"
)

type ConcurrentValue[T any] struct {
	value T
	mu    *sync.RWMutex
}

func NewConcurrentValue[T any](value T) *ConcurrentValue[T] {
	return &ConcurrentValue[T]{
		value: value,
		mu:    &sync.RWMutex{},
	}
}

func (t *ConcurrentValue[T]) Get() T {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.value
}

func (t *ConcurrentValue[T]) Set(value T) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.value = value
}
