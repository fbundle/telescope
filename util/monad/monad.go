package monad

type Monad[T any] struct {
	Chan  <-chan T
	Error error
}

func Unit[T any](vs ...T) Monad[T] {
	ch := make(chan T, 1)
	go func() {
		for _, v := range vs {
			ch <- v
		}
		close(ch)
	}()
	return Monad[T]{
		Chan:  ch,
		Error: nil,
	}
}

func Bind[T1 any, T2 any](m1 Monad[T1], f func(T1) Monad[T2]) Monad[T2] {
	ch := make(chan T2, 1)
	go func() {
		for v := range m1.Chan {
			m2 := f(v)
			for v2 := range m2.Chan {
				ch <- v2
			}
		}
		close(ch)
	}()
	return Monad[T2]{
		Chan:  ch,
		Error: nil,
	}
}
