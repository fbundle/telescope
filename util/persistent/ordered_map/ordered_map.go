package ordered_map

import "golang.org/x/exp/constraints"

type OrderedMap[K constraints.Ordered, V any] interface {
	Get(K) (V, bool)
	Set(K, V) OrderedMap[K, V]
	Del(K) OrderedMap[K, V]
	Iter(func(K, V) bool)
	Len() int
	Split(K) (OrderedMap[K, V], OrderedMap[K, V])
	Max() (K, V)
	Min() (K, V)
	Repr() map[K]V
}

func NewOrderedMap[K constraints.Ordered, V any]() OrderedMap[K, V] {
	return orderedMap[K, V]{
		comparableMap: NewMap[entry[K, V]](),
	}
}

type entry[K constraints.Ordered, V any] struct {
	key K
	val V
}

func (e entry[K, V]) Cmp(o entry[K, V]) int {
	switch {
	case e.key < o.key:
		return -1
	case e.key > o.key:
		return +1
	default:
		return 0
	}
}

type orderedMap[K constraints.Ordered, V any] struct {
	comparableMap Map[entry[K, V]]
}

func (o orderedMap[K, V]) Get(key K) (V, bool) {
	entryOut, ok := o.comparableMap.Get(entry[K, V]{key: key})
	return entryOut.(entry[K, V]).val, ok
}

func (o orderedMap[K, V]) Set(key K, val V) OrderedMap[K, V] {
	return orderedMap[K, V]{
		comparableMap: o.comparableMap.Set(entry[K, V]{key: key, val: val}),
	}
}

func (o orderedMap[K, V]) Del(key K) OrderedMap[K, V] {
	return orderedMap[K, V]{
		comparableMap: o.comparableMap.Del(entry[K, V]{key: key}),
	}
}

func (o orderedMap[K, V]) Iter(f func(K, V) bool) {
	o.comparableMap.Iter(func(entryOut entry[K, V]) bool {
		return f(entryOut.key, entryOut.val)
	})
}

func (o orderedMap[K, V]) Len() int {
	return int(o.comparableMap.Weight())
}

func (o orderedMap[K, V]) Split(key K) (OrderedMap[K, V], OrderedMap[K, V]) {
	m1, m2 := o.comparableMap.Split(entry[K, V]{key: key})
	return orderedMap[K, V]{comparableMap: m1}, orderedMap[K, V]{comparableMap: m2}
}

func (o orderedMap[K, V]) Max() (K, V) {
	entryOut := o.comparableMap.Max().(entry[K, V])
	return entryOut.key, entryOut.val
}

func (o orderedMap[K, V]) Min() (K, V) {
	entryOut := o.comparableMap.Min().(entry[K, V])
	return entryOut.key, entryOut.val
}

func (o orderedMap[K, V]) Repr() map[K]V {
	m := make(map[K]V)
	o.Iter(func(k K, v V) bool {
		m[k] = v
		return true
	})
	return m
}
