package ordered_map

import "golang.org/x/exp/constraints"

func NewOrderedMap[K constraints.Ordered, V any]() OrderedMap[K, V] {
	return OrderedMap[K, V]{
		Map: Empty[Entry[K, V]](),
	}
}

type Entry[K constraints.Ordered, V any] struct {
	Key K
	Val V
}

func (e Entry[K, V]) Cmp(o Entry[K, V]) int {
	switch {
	case e.Key < o.Key:
		return -1
	case e.Key > o.Key:
		return +1
	default:
		return 0
	}
}

type OrderedMap[K constraints.Ordered, V any] struct {
	Map Map[Entry[K, V]]
}

func (m OrderedMap[K, V]) Get(key K) (V, bool) {
	entryOut, ok := m.Map.Get(Entry[K, V]{Key: key})
	return entryOut.(Entry[K, V]).Val, ok
}

func (m OrderedMap[K, V]) Set(key K, val V) OrderedMap[K, V] {
	return OrderedMap[K, V]{
		Map: m.Map.Set(Entry[K, V]{Key: key, Val: val}),
	}
}

func (m OrderedMap[K, V]) Del(key K) OrderedMap[K, V] {
	return OrderedMap[K, V]{
		Map: m.Map.Del(Entry[K, V]{Key: key}),
	}
}

func (m OrderedMap[K, V]) Iter(f func(K, V) bool) {
	m.Map.Iter(func(entryOut Entry[K, V]) bool {
		return f(entryOut.Key, entryOut.Val)
	})
}

func (m OrderedMap[K, V]) Len() int {
	return int(m.Map.Weight())
}

func (m OrderedMap[K, V]) Split(key K) (OrderedMap[K, V], OrderedMap[K, V]) {
	m1, m2 := m.Map.Split(Entry[K, V]{Key: key})
	return OrderedMap[K, V]{Map: m1}, OrderedMap[K, V]{Map: m2}
}

func (m OrderedMap[K, V]) Max() (K, V) {
	entryOut := m.Map.Max().(Entry[K, V])
	return entryOut.Key, entryOut.Val
}

func (m OrderedMap[K, V]) Min() (K, V) {
	entryOut := m.Map.Min().(Entry[K, V])
	return entryOut.Key, entryOut.Val
}

func (m OrderedMap[K, V]) Repr() map[K]V {
	m1 := make(map[K]V)
	m.Iter(func(k K, v V) bool {
		m1[k] = v
		return true
	})
	return m1
}
