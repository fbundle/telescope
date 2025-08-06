package hist

import (
	"telescope/config"
)

func New[T any](t T) *Hist[T] {
	return &Hist[T]{
		LatestVersion: 0,
		Stack:         []T{t},
	}
}

type Hist[T any] struct {
	LatestVersion uint64 // current version
	Stack         []T    // all versions
}

func (h *Hist[T]) Update(modifier func(T) T) {
	next := modifier(h.Stack[h.LatestVersion])
	h.Stack = append(h.Stack[:h.LatestVersion+1], next)
	h.LatestVersion++
	if len(h.Stack) > config.Load().MAXSIZE_HISTORY {
		h.Stack = h.Stack[1:]
		h.LatestVersion--
	}
}

func (h *Hist[T]) Get() T {
	return h.Stack[h.LatestVersion]
}
func (h *Hist[T]) Undo() {
	if h.LatestVersion > 0 {
		h.LatestVersion--
	}
}

func (h *Hist[T]) Redo() {
	if int(h.LatestVersion) < len(h.Stack)-1 {
		h.LatestVersion++
	}
}
