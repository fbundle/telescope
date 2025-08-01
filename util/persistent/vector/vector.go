package vector

type Vector[T any] interface {
	Get(i int) T
	Set(i int, val T) Vector[T]
	Ins(i int, val T) Vector[T]
	Del(i int) Vector[T]
	Iter(func(i int, val T) bool)
	Len() int
	Height() int
	Split(i int) (Vector[T], Vector[T])
	Concat(other Vector[T]) Vector[T]
	Repr() []T
}

func New[T any]() Vector[T] {
	return vector[T]{node: nil}
}

type vector[T any] struct {
	node *node[T]
}

func (v vector[T]) Get(i int) T {
	return get(v.node, uint(i))
}

func (v vector[T]) Set(i int, val T) Vector[T] {
	return vector[T]{node: set(v.node, uint(i), val)}
}

func (v vector[T]) Ins(i int, val T) Vector[T] {
	return vector[T]{node: ins(v.node, uint(i), val)}
}

func (v vector[T]) Del(i int) Vector[T] {
	return vector[T]{node: del(v.node, uint(i))}
}

func (v vector[T]) Iter(f func(i int, val T) bool) {
	i := 0
	iter(v.node, func(val T) bool {
		ok := f(i, val)
		i++
		return ok
	})
}

func (v vector[T]) Len() int {
	return int(weight(v.node))
}
func (v vector[T]) Height() int {
	return int(height(v.node))
}
func (v vector[T]) Split(i int) (Vector[T], Vector[T]) {
	n1, n2 := split(v.node, uint(i))
	return vector[T]{node: n1}, vector[T]{node: n2}
}

func (v vector[T]) Concat(other Vector[T]) Vector[T] {
	n1, n2 := v.node, other.(*vector[T]).node
	n3 := merge(n1, n2)
	return vector[T]{node: n3}
}

func (v vector[T]) Repr() []T {
	buffer := make([]T, 0, v.Len())
	v.Iter(func(i int, val T) bool {
		buffer = append(buffer, val)
		return true
	})
	return buffer
}
