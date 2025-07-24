package ordered_map

import "golang.org/x/exp/constraints"

type OrderedMap[K constraints.Ordered, V any] interface {
	Get(K) (V, bool)
	Set(K, V) OrderedMap[K, V]
	Del(K) OrderedMap[K, V]
	Iter(func(K, V) bool)
	Weight() int
	Height() int
	Split(K) (OrderedMap[K, V], OrderedMap[K, V])
	Max() (K, V)
	Min() (K, V)
	Repr() map[K]V
}

func NewOrderedMap[K constraints.Ordered, V any]() OrderedMap[K, V] {
	return &orderedMap[K, V]{node: nil}
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
	node *node[entry[K, V]]
}

func (o *orderedMap[K, V]) Get(k K) (V, bool) {
	e, ok := get(o.node, entry[K, V]{key: k})
	return e.val, ok
}

func (o *orderedMap[K, V]) Set(key K, val V) OrderedMap[K, V] {
	return &orderedMap[K, V]{node: set(o.node, entry[K, V]{key: key, val: val})}
}

func (o *orderedMap[K, V]) Del(key K) OrderedMap[K, V] {
	return &orderedMap[K, V]{node: del(o.node, entry[K, V]{key: key})}
}

func (o *orderedMap[K, V]) Iter(f func(K, V) bool) {
	iter(o.node, func(e entry[K, V]) bool {
		return f(e.key, e.val)
	})
}

func (o *orderedMap[K, V]) Weight() int {
	return int(weight(o.node))
}

func (o *orderedMap[K, V]) Height() int {
	return int(height(o.node))
}
func (o *orderedMap[K, V]) Split(k K) (OrderedMap[K, V], OrderedMap[K, V]) {
	l, r := split(o.node, entry[K, V]{key: k})
	return &orderedMap[K, V]{node: l}, &orderedMap[K, V]{node: r}
}

func (o *orderedMap[K, V]) Max() (K, V) {
	e := getMaxEntry(o.node)
	return e.key, e.val
}
func (o *orderedMap[K, V]) Min() (K, V) {
	e := getMinEntry(o.node)
	return e.key, e.val
}

func (o *orderedMap[K, V]) Repr() map[K]V {
	m := make(map[K]V)
	o.Iter(func(k K, v V) bool {
		m[k] = v
		return true
	})
	return m
}
