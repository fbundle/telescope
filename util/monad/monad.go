package monad

type Monad[T any] struct {
	Chan  <-chan T
	Error error
}

func (m *Monad[T]) Unwrap() ([]T, error) {
	var vs []T
	for v := range m.Chan {
		vs = append(vs, v)
	}
	return vs, m.Error
}

func None[T any](err error) *Monad[T] {
	ch := make(chan T, 1)
	close(ch)
	m := &Monad[T]{
		Chan:  ch,
		Error: err,
	}
	return m
}

func Unit[T any](vs ...T) *Monad[T] {
	ch := make(chan T, 1)
	m := &Monad[T]{
		Chan:  ch,
		Error: nil,
	}
	go func() {
		for _, v := range vs {
			ch <- v
		}
		close(ch)
	}()
	return m
}

func Bind[T1 any, T2 any](m1 *Monad[T1], f func(T1) (*Monad[T2], error)) *Monad[T2] {
	ch := make(chan T2, 1)
	m := &Monad[T2]{
		Chan:  ch,
		Error: nil,
	}
	go func() {
		for v := range m1.Chan {
			m2, err := f(v)
			if err != nil {
				m.Error = err
				break
			}
			for v2 := range m2.Chan {
				ch <- v2
			}
			if err := m2.Error; err != nil {
				m.Error = err
				break
			}
		}
		close(ch)
	}()
	return m
}
