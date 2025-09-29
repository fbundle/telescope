package seq

func Empty[T any]() Seq[T] {
	return Seq[T]{node: nil}
}

type Seq[T any] struct {
	node *node[T]
}

func (s Seq[T]) Get(i int) T {
	return get(s.node, uint64(i))
}

func (s Seq[T]) Set(i int, val T) Seq[T] {
	return Seq[T]{node: set(s.node, uint64(i), val)}
}

func (s Seq[T]) Ins(i int, val T) Seq[T] {
	return Seq[T]{node: ins(s.node, uint64(i), val)}
}

func (s Seq[T]) Del(i int) Seq[T] {
	return Seq[T]{node: del(s.node, uint64(i))}
}

func (s Seq[T]) Iter(f func(i int, val T) bool) {
	i := 0
	iter(s.node, func(val T) bool {
		ok := f(i, val)
		i++
		return ok
	})
}

func (s Seq[T]) Len() int {
	return int(weight(s.node))
}

func (s Seq[T]) Split(i int) (Seq[T], Seq[T]) {
	n1, n2 := split(s.node, uint64(i))
	return Seq[T]{node: n1}, Seq[T]{node: n2}
}

func (s Seq[T]) Merge(other Seq[T]) Seq[T] {
	n1, n2 := s.node, other.node
	n3 := merge(n1, n2)
	return Seq[T]{node: n3}
}

func (s Seq[T]) Repr() []T {
	buffer := make([]T, 0, s.Len())
	s.Iter(func(i int, val T) bool {
		buffer = append(buffer, val)
		return true
	})
	return buffer
}
