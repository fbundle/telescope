package iterator

type Iterator[T any] = func(func(T) bool)
