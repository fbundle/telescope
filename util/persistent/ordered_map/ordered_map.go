package ordered_map

import (
	"cmp"
)

func EmptyOrderedMap[K cmp.Ordered, V any]() OrderedMap[K, V] {
	return OrderedMap[K, V]{
		comparableMap: EmptyComparableMap[Entry[K, V]](),
	}
}

type Entry[K cmp.Ordered, V any] struct {
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

type OrderedMap[K cmp.Ordered, V any] struct {
	comparableMap Map[Entry[K, V]]
}

func (m OrderedMap[K, V]) Get(key K) (V, bool) {
	entryOut, ok := m.comparableMap.Get(Entry[K, V]{Key: key})
	return entryOut.(Entry[K, V]).Val, ok
}

func (m OrderedMap[K, V]) Set(key K, val V) OrderedMap[K, V] {
	return OrderedMap[K, V]{
		comparableMap: m.comparableMap.Set(Entry[K, V]{Key: key, Val: val}),
	}
}

func (m OrderedMap[K, V]) Del(key K) OrderedMap[K, V] {
	return OrderedMap[K, V]{
		comparableMap: m.comparableMap.Del(Entry[K, V]{Key: key}),
	}
}

func (m OrderedMap[K, V]) Iter(f func(K, V) bool) {
	m.comparableMap.Iter(func(entryOut Entry[K, V]) bool {
		return f(entryOut.Key, entryOut.Val)
	})
}

func (m OrderedMap[K, V]) Len() int {
	return int(m.comparableMap.Len())
}

func (m OrderedMap[K, V]) Split(key K) (OrderedMap[K, V], OrderedMap[K, V]) {
	m1, m2 := m.comparableMap.Split(Entry[K, V]{Key: key})
	return OrderedMap[K, V]{comparableMap: m1}, OrderedMap[K, V]{comparableMap: m2}
}

func (m OrderedMap[K, V]) Max() (K, V) {
	entryOut := m.comparableMap.Max().(Entry[K, V])
	return entryOut.Key, entryOut.Val
}

func (m OrderedMap[K, V]) Min() (K, V) {
	entryOut := m.comparableMap.Min().(Entry[K, V])
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
