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

func New[T any]() *WeightBalancedTree[T] {
	return &WeightBalancedTree[T]{node: nil}
}

type WeightBalancedTree[T any] struct {
	node *node[T]
}

func (v *WeightBalancedTree[T]) Get(i int) T {
	return get(v.node, uint(i))
}

func (v *WeightBalancedTree[T]) Set(i int, val T) Vector[T] {
	return &WeightBalancedTree[T]{node: set(v.node, uint(i), val)}
}

func (v *WeightBalancedTree[T]) Ins(i int, val T) Vector[T] {
	return &WeightBalancedTree[T]{node: ins(v.node, uint(i), val)}
}

func (v *WeightBalancedTree[T]) Del(i int) Vector[T] {
	return &WeightBalancedTree[T]{node: del(v.node, uint(i))}
}

func (v *WeightBalancedTree[T]) Iter(f func(i int, val T) bool) {
	i := 0
	iter(v.node, func(val T) bool {
		ok := f(i, val)
		i++
		return ok
	})
}

func (v *WeightBalancedTree[T]) Len() int {
	return int(weight(v.node))
}
func (v *WeightBalancedTree[T]) Height() int {
	return int(height(v.node))
}
func (v *WeightBalancedTree[T]) Split(i int) (Vector[T], Vector[T]) {
	n1, n2 := split(v.node, uint(i))
	return &WeightBalancedTree[T]{node: n1}, &WeightBalancedTree[T]{node: n2}
}

func (v *WeightBalancedTree[T]) Concat(other Vector[T]) Vector[T] {
	n1, n2 := v.node, other.(*WeightBalancedTree[T]).node
	n3 := merge(n1, n2)
	return &WeightBalancedTree[T]{node: n3}
}

func (v *WeightBalancedTree[T]) Repr() []T {
	buffer := make([]T, 0, v.Len())
	v.Iter(func(i int, val T) bool {
		buffer = append(buffer, val)
		return true
	})
	return buffer
}
