package buffer

import "io"

type Slice struct {
	Offset int
	Size   int
}

func (s Slice) Repr(b Buffer) []byte {
	buf := make([]byte, s.Size)
	_, err := b.ReadAt(buf, int64(s.Offset))
	if err != nil && err != io.EOF {
		panic(err)
	}
	return buf
}
