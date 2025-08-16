package atomic_util

import (
	"sync"
)

func NewOnce[T any]() *Once[T] {
	return &Once[T]{
		mu: sync.Mutex{},
	}
}

type Once[T any] struct {
	mu   sync.Mutex
	full bool
	v    T
}

func (o *Once[T]) LoadOrStore(store func() T) T {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.full {
		o.v = store()
	}
	return o.v
}
