package sync_util

import "sync"

// just a type-safe sync.Map

type Map[K any, V any] struct {
	m sync.Map
}

func (sm *Map[K, V]) Clear() {
	sm.m.Clear()
}

func (sm *Map[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return sm.m.CompareAndDelete(key, old)
}

func (sm *Map[K, V]) CompareAndSwap(key K, old V, new V) (swapped bool) {
	return sm.m.CompareAndSwap(key, old, new)
}

func (sm *Map[K, V]) Delete(key K) {
	sm.m.Delete(key)
}

func zero[T any]() T {
	var z T
	return z
}

func (sm *Map[K, V]) Load(key K) (value V, ok bool) {
	val, ok := sm.m.Load(key)
	if !ok {
		return zero[V](), false
	}
	return val.(V), true
}

func (sm *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	val, loaded := sm.m.LoadAndDelete(key)
	if !loaded {
		return zero[V](), false
	}
	return val.(V), true
}

func (sm *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	act, loaded := sm.m.LoadOrStore(key, value)
	return act.(V), loaded
}

func (sm *Map[K, V]) Range(f func(key K, value V) bool) {
	sm.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

func (sm *Map[K, V]) Store(key K, value V) {
	sm.m.Store(key, value)
}

func (sm *Map[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	prev, loaded := sm.m.Swap(key, value)
	if !loaded {
		return zero[V](), false
	}
	return prev.(V), loaded
}
