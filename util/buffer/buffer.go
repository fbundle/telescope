package buffer

import "io"

type Buffer interface {
	Len() int
	At(i int) byte
}

func NewMemBuffer(b []byte) Buffer {
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

func (m *memBuffer) ReadAt(b []byte, i int64) (n int, err error) {
	if i < 0 || i >= int64(len(m.b)) {
		return 0, io.EOF
	}
	n = copy(b, m.b[i:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}
