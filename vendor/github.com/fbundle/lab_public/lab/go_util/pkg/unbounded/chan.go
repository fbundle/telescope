package unbounded

import (
	"context"
	"fmt"
)

const (
	bufferSize = 1024
)

func New[T any](ctx context.Context) *Chan[T] {
	ch := &Chan[T]{
		inCh:  make(chan T, bufferSize),
		outCh: make(chan T, bufferSize),
	}
	go ch.loop(ctx)
	return ch
}

type Chan[T any] struct {
	inCh  chan T
	outCh chan T
}

func (ch *Chan[T]) InChan() chan<- T {
	return ch.inCh
}
func (ch *Chan[T]) OutChan() <-chan T {
	return ch.outCh
}

func (ch *Chan[T]) loop(ctx context.Context) {
	defer close(ch.outCh)

	closed := false
	queue := make([]T, 0)
	for {
		fmt.Println(len(queue))
		if closed {
			for len(queue) > 0 {
				select {
				case <-ctx.Done():
					return
				case ch.outCh <- queue[0]:
					queue = queue[1:]
				}
			}
			return
		}
		if len(queue) == 0 {
			select {
			case <-ctx.Done():
				return
			case o, ok := <-ch.inCh:
				if !ok { // inCh closed
					closed = true
				} else {
					queue = append(queue, o)
				}
			}
		} else {
			select {
			case <-ctx.Done():
				return
			case o, ok := <-ch.inCh:
				if !ok { // inCh closed
					closed = true
				} else {
					queue = append(queue, o)
				}
			case ch.outCh <- queue[0]:
				queue = queue[1:]
			}
		}
	}
}
