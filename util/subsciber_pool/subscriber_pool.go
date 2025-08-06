package subsciber_pool

type Pool[T any] struct {
	handlerMap map[uint64]T
	lastKey    uint64
}

func New[T any]() *Pool[T] {
	return &Pool[T]{
		handlerMap: make(map[uint64]T),
		lastKey:    0,
	}
}

func (p *Pool[T]) Subscribe(a T) uint64 {
	p.lastKey++
	p.handlerMap[p.lastKey] = a
	return p.lastKey
}

func (p *Pool[T]) Unsubscribe(key uint64) {
	delete(p.handlerMap, key)
}

func (p *Pool[T]) Iter(f func(k uint64, v T) bool) {
	for k, v := range p.handlerMap {
		if !f(k, v) {
			break
		}
	}
}
