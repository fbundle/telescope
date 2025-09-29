package ordered_map

func EmptyComparableMap[T Comparable[T]]() Map[T] {
	return Map[T]{
		node: nil,
	}
}

type Map[T Comparable[T]] struct {
	node *node[T]
}

func (m Map[T]) Get(entryIn T) (Comparable[T], bool) {
	entryOut, ok := get(m.node, entryIn)
	return entryOut, ok
}

func (m Map[T]) Set(entryIn T) Map[T] {
	return Map[T]{
		node: set(m.node, entryIn),
	}
}

func (m Map[T]) Del(entryIn T) Map[T] {
	return Map[T]{
		node: del(m.node, entryIn),
	}
}

func (m Map[T]) Iter(f func(entryOut T) bool) {
	iter(m.node, f)
}

func (m Map[T]) Len() uint64 {
	return weight(m.node)
}

func (m Map[T]) Split(entryIn T) (Map[T], Map[T]) {
	n1, n2 := split(m.node, entryIn)
	return Map[T]{node: n1}, Map[T]{node: n2}
}

func (m Map[T]) Max() Comparable[T] {
	return getMaxEntry(m.node)
}

func (m Map[T]) Min() Comparable[T] {
	return getMinEntry(m.node)
}

func (m Map[T]) Repr() []Comparable[T] {
	buf := make([]Comparable[T], 0, m.Len())
	for entryOut := range m.Iter {
		buf = append(buf, entryOut)
	}
	return buf
}
