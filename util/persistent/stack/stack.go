package stack

func Empty[T any]() Stack[T] {
	return Stack[T]{node: nil}
}

type Stack[T any] struct {
	node *node[T]
}

func (s Stack[T]) Peek() T {
	return s.node.value
}

func (s Stack[T]) Pop() Stack[T] {
	return Stack[T]{
		node: s.node.next,
	}
}

func (s Stack[T]) Depth() int {
	return int(depth(s.node))
}

func (s Stack[T]) Push(v T) Stack[T] {
	return Stack[T]{
		node: newNode(v, s.node),
	}
}

func (s Stack[T]) Iter(f func(i int, val T) bool) {
	length := s.Depth()
	if length == 0 {
		return
	}
	if ok := f(length-1, s.node.value); !ok {
		return
	}
	s.Pop().Iter(f)
}

func (s Stack[T]) Repr() []T {
	buffer := make([]T, 0, s.Depth())
	s.Iter(func(i int, val T) bool {
		buffer = append(buffer, val)
		return true
	})
	return buffer
}
