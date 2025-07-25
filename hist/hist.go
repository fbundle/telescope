package hist

type Hist[T any] interface {
	Update(modifier func(T) T)
	Get() T
	Undo()
	Redo()
}

func New[T any](t T) Hist[T] {
	return &hist[T]{
		currentVersion: 0,
		versionList:    []T{t},
	}
}

type hist[T any] struct {
	currentVersion int
	versionList    []T
}

func (h *hist[T]) Update(modifier func(T) T) {
	h.versionList = h.versionList[:h.currentVersion+1]
	next := modifier(h.versionList[h.currentVersion])
	h.versionList = append(h.versionList, next)
	h.currentVersion++
}

func (h *hist[T]) Get() T {
	return h.versionList[h.currentVersion]
}
func (h *hist[T]) Undo() {
	if h.currentVersion > 0 {
		h.currentVersion--
	}
}

func (h *hist[T]) Redo() {
	if h.currentVersion < len(h.versionList)-1 {
		h.currentVersion++
	}
}
