package mutex

import "sync"

type Data[T any] struct {
	mu   sync.Mutex
	Data T
}

func NewData[T any](data T) *Data[T] {
	return &Data[T]{Data: data}
}

func (m *Data[T]) Lock() {
	m.mu.Lock()
}

func (m *Data[T]) Unlock() {
	m.mu.Unlock()
}
