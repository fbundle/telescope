package seq

func (s Seq[T]) Front() T {
	return s.Get(0)
}

func (s Seq[T]) Back() T {
	return s.Get(s.Len() - 1)
}

func (s Seq[T]) PushFront(vals ...T) Seq[T] {
	for i := len(vals) - 1; i >= 0; i-- {
		s = s.Ins(0, vals[i])
	}
	return s
}
func (s Seq[T]) PushBack(vals ...T) Seq[T] {
	for i := 0; i < len(vals); i++ {
		s = s.Ins(s.Len(), vals[i])
	}
	return s
}

func (s Seq[T]) PopFront() Seq[T] {
	return s.Del(0)
}

func (s Seq[T]) PopBack() Seq[T] {
	return s.Del(s.Len() - 1)
}

func (s Seq[T]) IndexOf(pred func(T) bool) int {
	index := -1
	for i, val := range s.Iter {
		if pred(val) {
			index = i
			break
		}
	}
	return index
}
func (s Seq[T]) Contains(pred func(T) bool) bool {
	return s.IndexOf(pred) >= 0
}

func (s Seq[T]) Slice(beg int, end int) Seq[T] {
	if beg > end {
		panic("slice out of range")
	}
	s, _ = s.Split(end)
	_, s = s.Split(beg)
	return s
}

func Merge[T any](ss ...Seq[T]) Seq[T] {
	if len(ss) == 0 {
		return Empty[T]()
	}
	s := ss[0]
	for i := 1; i < len(ss); i++ {
		s = s.Merge(ss[i])
	}
	return s
}
