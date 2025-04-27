package collection

import (
	"fmt"
	"sync"
)

type ConcurrentList[T any] struct {
	list []T
	mu   *sync.RWMutex
}

func NewConcurrentList[T any]() *ConcurrentList[T] {
	return &ConcurrentList[T]{
		mu: &sync.RWMutex{},
	}
}

func (t *ConcurrentList[T]) Get(index int) (T, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if index >= len(t.list) {
		var null T
		return null, fmt.Errorf("index out of bound")
	}
	return t.list[index], nil
}

func (t *ConcurrentList[T]) Set(index int, value T) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if index >= len(t.list) {
		return fmt.Errorf("index out of bound")
	}
	t.list[index] = value
	return nil
}

func (t *ConcurrentList[T]) Append(value T) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.list = append(t.list, value)
}

// Don't need lock, bottleneck
func (t *ConcurrentList[T]) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.list)
}

func (t *ConcurrentList[T]) RemoveFirst() T {
	t.mu.Lock()
	defer t.mu.Unlock()
	var value T
	if len(t.list) > 0 {
		value = t.list[0]
		t.list = t.list[1:]
	}
	return value
}

func (t *ConcurrentList[T]) Range(start int, end int) []T {
	var data []T
	t.mu.RLock()
	defer t.mu.RUnlock()
	end = min(end, len(t.list))
	if start >= end {
		return data
	}
	data = append(data, t.list[start:end]...)
	return data
}
