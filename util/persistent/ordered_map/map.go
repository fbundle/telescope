package ordered_map

type Map[T Comparable[T]] interface {
	Get(T) (Comparable[T], bool)
	Set(T) Map[T]
	Del(T) Map[T]
	Iter(func(T) bool)
	Weight() uint
	Split(T) (Map[T], Map[T])
	Max() Comparable[T]
	Min() Comparable[T]
	Repr() []Comparable[T]
}

func NewMap[T Comparable[T]]() WBT[T] {
	return WBT[T]{
		node: nil,
	}
}

type WBT[T Comparable[T]] struct {
	node *node[T]
}

func (o WBT[T]) Get(entryIn T) (Comparable[T], bool) {
	entryOut, ok := get(o.node, entryIn)
	return entryOut, ok
}

func (o WBT[T]) Set(entryIn T) Map[T] {
	return WBT[T]{
		node: set(o.node, entryIn),
	}
}

func (o WBT[T]) Del(entryIn T) Map[T] {
	return WBT[T]{
		node: del(o.node, entryIn),
	}
}

func (o WBT[T]) Iter(f func(entryOut T) bool) {
	iter(o.node, f)
}

func (o WBT[T]) Weight() uint {
	return weight(o.node)
}

func (o WBT[T]) Height() uint {
	return height(o.node)
}

func (o WBT[T]) Split(entryIn T) (Map[T], Map[T]) {
	n1, n2 := split(o.node, entryIn)
	return WBT[T]{node: n1}, WBT[T]{node: n2}
}

func (o WBT[T]) Max() Comparable[T] {
	return getMaxEntry(o.node)
}

func (o WBT[T]) Min() Comparable[T] {
	return getMinEntry(o.node)
}

func (o WBT[T]) Repr() []Comparable[T] {
	buf := make([]Comparable[T], 0, o.Weight())
	for entryOut := range o.Iter {
		buf = append(buf, entryOut)
	}
	return buf
}
