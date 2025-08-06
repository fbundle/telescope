package hist

import (
	"telescope/config"
)

func New[T any](t T) *Hist[T] {
	return &Hist[T]{
		i:  0,
		ts: []T{t},
	}
}

type Hist[T any] struct {
	i  int // current version
	ts []T // all versions
}

func (h *Hist[T]) Update(modifier func(T) T) {
	next := modifier(h.ts[h.i])
	h.ts = append(h.ts[:h.i+1], next)
	h.i++
	if len(h.ts) > config.Load().MAXSIZE_HISTORY {
		h.ts = h.ts[1:]
		h.i--
	}
}

func (h *Hist[T]) Get() T {
	return h.ts[h.i]
}
func (h *Hist[T]) Undo() {
	if h.i > 0 {
		h.i--
	}
}

func (h *Hist[T]) Redo() {
	if h.i < len(h.ts)-1 {
		h.i++
	}
}
