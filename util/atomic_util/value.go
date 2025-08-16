package atomic_util

import (
	"sync/atomic"
)

func NewValue[T any]() *Value[T] {
	return &Value[T]{
		v: atomic.Value{},
	}
}

// just type-safe atomic value

type Value[T any] struct {
	v atomic.Value
}

func unwrap[T any](o any) (val T, ok bool) {
	if o == nil {
		return val, false
	}
	return o.(T), true
}

func (av *Value[T]) Load() (val T, ok bool) {
	return unwrap[T](av.v.Load())
}
func (av *Value[T]) Store(val T) {
	av.v.Store(val)
}
func (av *Value[T]) Swap(new T) (old T, ok bool) {
	return unwrap[T](av.v.Swap(new))
}
func (av *Value[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return av.v.CompareAndSwap(old, new)
}
