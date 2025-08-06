package future

import "context"

type Future[T any] struct {
	Error  error
	Value  T
	getCtx context.Context
}

func (f *Future[T]) Done() <-chan struct{} {
	return f.getCtx.Done()
}

func New[T any](ctx context.Context, get func(context.Context) (T, error)) *Future[T] {
	getCtx, cancel := context.WithCancel(context.Background())
	o := &Future[T]{
		getCtx: getCtx,
	}
	go func() {
		defer cancel()
		o.Value, o.Error = get(ctx)
	}()
	return o
}
