package collection

import (
	"fmt"
	"sync"
)

type ConcurrentMap[K comparable, V any] struct {
	kv map[K]V
	mu *sync.RWMutex
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		kv: make(map[K]V),
		mu: &sync.RWMutex{},
	}
}

func (t *ConcurrentMap[K, V]) Get(key K) (V, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if _, ok := t.kv[key]; !ok {
		var null V
		return null, fmt.Errorf("key not found")
	}
	return t.kv[key], nil
}

func (t *ConcurrentMap[K, V]) Set(key K, value V) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.kv[key] = value
}

func (t *ConcurrentMap[K, V]) Delete(key K) (V, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.kv[key]; !ok {
		var null V
		return null, fmt.Errorf("key not found")
	}
	cop := t.kv[key]
	delete(t.kv, key)
	return cop, nil
}

func (t *ConcurrentMap[K, V]) Values() []V {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var values []V
	for _, v := range t.kv {
		values = append(values, v)
	}
	return values
}

func (t *ConcurrentMap[K, V]) Keys() []K {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var keys []K
	for k := range t.kv {
		keys = append(keys, k)
	}
	return keys
}

func (t *ConcurrentMap[K, V]) Entries() ([]K, []V) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var values []V
	var keys []K
	for k, v := range t.kv {
		values = append(values, v)
		keys = append(keys, k)
	}
	return keys, values
}
