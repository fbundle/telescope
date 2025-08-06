package subsciber_pool

import (
	"sync"
)

type Pool[T any] struct {
	mu         sync.RWMutex
	handlerMap map[uint64]T
	lastKey    uint64
}

func New[T any]() *Pool[T] {
	return &Pool[T]{
		mu:         sync.RWMutex{},
		handlerMap: make(map[uint64]T),
		lastKey:    0,
	}
}

func (p *Pool[T]) Subscribe(a T) uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lastKey++
	p.handlerMap[p.lastKey] = a
	return p.lastKey
}

func (p *Pool[T]) Unsubscribe(key uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.handlerMap, key)
}

func (p *Pool[T]) Iter(f func(k uint64, v T) bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for k, v := range p.handlerMap {
		if !f(k, v) {
			break
		}
	}
}
