package stack

type node[T any] struct {
	value T
	depth uint
	next  *node[T]
}

func depth[T any](n *node[T]) uint {
	if n == nil {
		return 0
	}
	return n.depth
}

func newNode[T any](value T, next *node[T]) *node[T] {
	return &node[T]{
		value: value,
		depth: depth(next) + 1,
		next:  next,
	}
}
