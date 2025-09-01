package buffer

type Reader interface {
	Len() int
	At(i int) byte
}

type SliceReader struct {
	reader Reader
	beg    int
	len    int
}

func (s SliceReader) Len() int {
	return s.len
}

func (s SliceReader) At(i int) byte {
	return s.reader.At(i + s.beg)
}

func Slice(reader Reader, beg int, end int) Reader {
	if r, ok := reader.(SliceReader); ok {
		return SliceReader{
			reader: r.reader,
			beg:    r.beg + beg,
			len:    end - beg,
		}
	} else {
		return SliceReader{
			reader: reader,
			beg:    beg,
			len:    end - beg,
		}
	}
}

func NewMemReader(b []byte) Reader {
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
