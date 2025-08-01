package vector

type Vector[T any] interface {
	Get(i int) T
	Set(i int, val T) Vector[T]
	Ins(i int, val T) Vector[T]
	Del(i int) Vector[T]
	Iter(func(i int, val T) bool)
	Len() int
	Split(i int) (Vector[T], Vector[T])
	Concat(other Vector[T]) Vector[T]
	Repr() []T
}

func New[T any]() WBT[T] {
	return WBT[T]{node: nil}
}

type WBT[T any] struct {
	node *node[T]
}

func (t WBT[T]) Get(i int) T {
	return get(t.node, uint(i))
}

func (t WBT[T]) Set(i int, val T) Vector[T] {
	return WBT[T]{node: set(t.node, uint(i), val)}
}

func (t WBT[T]) Ins(i int, val T) Vector[T] {
	return WBT[T]{node: ins(t.node, uint(i), val)}
}

func (t WBT[T]) Del(i int) Vector[T] {
	return WBT[T]{node: del(t.node, uint(i))}
}

func (t WBT[T]) Iter(f func(i int, val T) bool) {
	i := 0
	iter(t.node, func(val T) bool {
		ok := f(i, val)
		i++
		return ok
	})
}

func (t WBT[T]) Len() int {
	return int(weight(t.node))
}
func (t WBT[T]) Height() int {
	return int(height(t.node))
}
func (t WBT[T]) Split(i int) (Vector[T], Vector[T]) {
	n1, n2 := split(t.node, uint(i))
	return WBT[T]{node: n1}, WBT[T]{node: n2}
}

func (t WBT[T]) Concat(other Vector[T]) Vector[T] {
	n1, n2 := t.node, other.(*WBT[T]).node
	n3 := merge(n1, n2)
	return WBT[T]{node: n3}
}

func (t WBT[T]) Repr() []T {
	buffer := make([]T, 0, t.Len())
	t.Iter(func(i int, val T) bool {
		buffer = append(buffer, val)
		return true
	})
	return buffer
}
