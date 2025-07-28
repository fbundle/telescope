package bytes

import "io"

type Array interface {
	Len() int
	At(i int) byte
	ReadAt(b []byte, i int64) (n int, err error)
}

type memArray struct {
	b []byte
}

func (m *memArray) Len() int {
	return len(m.b)
}

func (m *memArray) At(i int) byte {
	return m.b[i]
}

func (m *memArray) ReadAt(b []byte, i int64) (n int, err error) {
	if i < 0 || i >= int64(len(m.b)) {
		return 0, io.EOF
	}
	n = copy(b, m.b[i:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

func NewMemArray(b []byte) Array {
	return &memArray{b: b}
}
