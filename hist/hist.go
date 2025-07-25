package hist

type Hist[T any] interface {
	Update(modifier func(T) T)
	Get() T
	Undo()
	Redo()
}

func New[T any](t T) Hist[T] {
	return &hist[T]{
		i:  0,
		ts: []T{t},
	}
}

type hist[T any] struct {
	i  int // current version
	ts []T // all versions
}

func (h *hist[T]) Update(modifier func(T) T) {
	next := modifier(h.ts[h.i])
	h.ts = append(h.ts[:h.i+1], next)
	h.i++
}

func (h *hist[T]) Get() T {
	return h.ts[h.i]
}
func (h *hist[T]) Undo() {
	if h.i > 0 {
		h.i--
	}
}

func (h *hist[T]) Redo() {
	if h.i < len(h.ts)-1 {
		h.i++
	}
}
