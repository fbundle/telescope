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

func NewMap[T Comparable[T]]() Map[T] {
	return wbt[T]{
		node: nil,
	}
}

type wbt[T Comparable[T]] struct {
	node *node[T]
}

func (o wbt[T]) Get(entryIn T) (Comparable[T], bool) {
	entryOut, ok := get(o.node, entryIn)
	return entryOut, ok
}

func (o wbt[T]) Set(entryIn T) Map[T] {
	return wbt[T]{
		node: set(o.node, entryIn),
	}
}

func (o wbt[T]) Del(entryIn T) Map[T] {
	return wbt[T]{
		node: del(o.node, entryIn),
	}
}

func (o wbt[T]) Iter(f func(entryOut T) bool) {
	iter(o.node, f)
}

func (o wbt[T]) Weight() uint {
	return weight(o.node)
}

func (o wbt[T]) Height() uint {
	return height(o.node)
}

func (o wbt[T]) Split(entryIn T) (Map[T], Map[T]) {
	n1, n2 := split(o.node, entryIn)
	return wbt[T]{node: n1}, wbt[T]{node: n2}
}

func (o wbt[T]) Max() Comparable[T] {
	return getMaxEntry(o.node)
}

func (o wbt[T]) Min() Comparable[T] {
	return getMinEntry(o.node)
}

func (o wbt[T]) Repr() []Comparable[T] {
	buf := make([]Comparable[T], 0, o.Weight())
	for entryOut := range o.Iter {
		buf = append(buf, entryOut)
	}
	return buf
}
