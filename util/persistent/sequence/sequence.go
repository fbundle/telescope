package sequence

type Seq[T any] interface {
	Get(i int) T
	Set(i int, val T) Seq[T]
	Ins(i int, val T) Seq[T]
	Del(i int) Seq[T]
	Iter(func(i int, val T) bool)
	Len() int
	Split(i int) (Seq[T], Seq[T])
	Concat(other Seq[T]) Seq[T]
	Repr() []T
}

func New[T any]() Seq[T] {
	return wbt[T]{node: nil}
}

type wbt[T any] struct {
	node *node[T]
}

func (t wbt[T]) Get(i int) T {
	return get(t.node, uint(i))
}

func (t wbt[T]) Set(i int, val T) Seq[T] {
	return wbt[T]{node: set(t.node, uint(i), val)}
}

func (t wbt[T]) Ins(i int, val T) Seq[T] {
	return wbt[T]{node: ins(t.node, uint(i), val)}
}

func (t wbt[T]) Del(i int) Seq[T] {
	return wbt[T]{node: del(t.node, uint(i))}
}

func (t wbt[T]) Iter(f func(i int, val T) bool) {
	i := 0
	iter(t.node, func(val T) bool {
		ok := f(i, val)
		i++
		return ok
	})
}

func (t wbt[T]) Len() int {
	return int(weight(t.node))
}
func (t wbt[T]) Height() int {
	return int(height(t.node))
}
func (t wbt[T]) Split(i int) (Seq[T], Seq[T]) {
	n1, n2 := split(t.node, uint(i))
	return wbt[T]{node: n1}, wbt[T]{node: n2}
}

func (t wbt[T]) Concat(other Seq[T]) Seq[T] {
	n1, n2 := t.node, other.(*wbt[T]).node
	n3 := merge(n1, n2)
	return wbt[T]{node: n3}
}

func (t wbt[T]) Repr() []T {
	buffer := make([]T, 0, t.Len())
	t.Iter(func(i int, val T) bool {
		buffer = append(buffer, val)
		return true
	})
	return buffer
}
