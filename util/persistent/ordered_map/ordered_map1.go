package ordered_map

type OrderedMap1[T Comparable[T]] interface {
	Get(T) T
	Set(T) OrderedMap1[T]
	Del(T) OrderedMap1[T]
	Iter(func(T) bool)
	Weight() int
	Height() int
	Split(T) (OrderedMap1[T], OrderedMap1[T])
	Max() T
	Min() T
	Repr() []T
}

func NewOrderedMap1[T Comparable[T]]() OrderedMap1[T] {
	return &orderedMap1[T]{
		node: nil,
	}
}

type orderedMap1[T Comparable[T]] struct {
	node *node[T]
}

func (o orderedMap1[T]) Get(t T) T {
	//TODO implement me
	panic("implement me")
}

func (o orderedMap1[T]) Set(t T) OrderedMap1[T] {
	//TODO implement me
	panic("implement me")
}

func (o orderedMap1[T]) Del(t T) OrderedMap1[T] {
	//TODO implement me
	panic("implement me")
}

func (o orderedMap1[T]) Iter(f func(T) bool) {
	//TODO implement me
	panic("implement me")
}

func (o orderedMap1[T]) Weight() int {
	//TODO implement me
	panic("implement me")
}

func (o orderedMap1[T]) Height() int {
	//TODO implement me
	panic("implement me")
}

func (o orderedMap1[T]) Split(t T) (OrderedMap1[T], OrderedMap1[T]) {
	//TODO implement me
	panic("implement me")
}

func (o orderedMap1[T]) Max() T {
	//TODO implement me
	panic("implement me")
}

func (o orderedMap1[T]) Min() T {
	//TODO implement me
	panic("implement me")
}

func (o orderedMap1[T]) Repr() []T {
	//TODO implement me
	panic("implement me")
}
