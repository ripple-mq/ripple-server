package collection

// Queue[T] holds thread safe queue
type ConcurrentQueue[T any] struct {
	q *ConcurrentList[T]
}

func NewConcurrentQueue[T any]() *ConcurrentQueue[T] {
	return &ConcurrentQueue[T]{
		q: NewConcurrentList[T](),
	}
}

func (t *ConcurrentQueue[T]) Size() int {
	return t.q.Size()
}

func (t *ConcurrentQueue[T]) IsEmpty() bool {
	return t.q.Size() == 0
}

func (t *ConcurrentQueue[T]) Push(value T) {
	t.q.Append(value)
}

func (t *ConcurrentQueue[T]) Poll() T {
	return t.q.RemoveFirst()
}

func (t *ConcurrentQueue[T]) SubArray(start int, end int) []T {
	return t.q.Range(start, end)
}
