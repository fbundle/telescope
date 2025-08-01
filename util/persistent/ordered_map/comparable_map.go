package ordered_map

type ComparableMap[T Comparable[T]] interface {
	Get(T) (Comparable[T], bool)
	Set(T) ComparableMap[T]
	Del(T) ComparableMap[T]
	Iter(func(T) bool)
	Weight() uint
	Height() uint
	Split(T) (ComparableMap[T], ComparableMap[T])
	Max() Comparable[T]
	Min() Comparable[T]
	Repr() []Comparable[T]
}

func NewComparableMap[T Comparable[T]]() ComparableMap[T] {
	return comparableMap[T]{
		node: nil,
	}
}

type comparableMap[T Comparable[T]] struct {
	node *node[T]
}

func (o comparableMap[T]) Get(entryIn T) (Comparable[T], bool) {
	entryOut, ok := get(o.node, entryIn)
	return entryOut, ok
}

func (o comparableMap[T]) Set(entryIn T) ComparableMap[T] {
	return comparableMap[T]{
		node: set(o.node, entryIn),
	}
}

func (o comparableMap[T]) Del(entryIn T) ComparableMap[T] {
	return comparableMap[T]{
		node: del(o.node, entryIn),
	}
}

func (o comparableMap[T]) Iter(f func(entryOut T) bool) {
	iter(o.node, f)
}

func (o comparableMap[T]) Weight() uint {
	return weight(o.node)
}

func (o comparableMap[T]) Height() uint {
	return height(o.node)
}

func (o comparableMap[T]) Split(entryIn T) (ComparableMap[T], ComparableMap[T]) {
	n1, n2 := split(o.node, entryIn)
	return comparableMap[T]{node: n1}, comparableMap[T]{node: n2}
}

func (o comparableMap[T]) Max() Comparable[T] {
	return getMaxEntry(o.node)
}

func (o comparableMap[T]) Min() Comparable[T] {
	return getMinEntry(o.node)
}

func (o comparableMap[T]) Repr() []Comparable[T] {
	buf := make([]Comparable[T], 0, o.Weight())
	for entryOut := range o.Iter {
		buf = append(buf, entryOut)
	}
	return buf
}
