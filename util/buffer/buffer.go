package buffer

type Reader interface {
	Len() int
	At(i int) byte
}

func NewMemBuffer(b []byte) Reader {
	return &memBuffer{b: b}
}

type memBuffer struct {
	b []byte
}

func (m *memBuffer) Len() int {
	return len(m.b)
}

func (m *memBuffer) At(i int) byte {
	return m.b[i]
}
