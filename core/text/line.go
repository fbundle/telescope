package text

import (
	"io"
	"telescope/util/buffer"
	"telescope/util/side_channel"
)

// Line - if offset >= 0, this is a file else this is a []rune buffer
type Line struct {
	offset int
	size   int
	data   *[]rune
}

func makeLineFromData(data []rune) Line {
	return Line{
		offset: -1,
		size:   -1,
		data:   &data,
	}
}

func makeLineFromFile(offset int, size int) Line {
	return Line{
		offset: offset,
		size:   size,
		data:   nil,
	}
}

func (l Line) Size() int {
	if l.offset < 0 {
		return len(*l.data)
	} else {
		return l.size
	}
}

func (l Line) Repr(r buffer.Buffer) []rune {
	if l.offset < 0 {
		// in-memory
		return *l.data
	} else {
		// from file
		buf := make([]byte, l.size)
		_, err := r.ReadAt(buf, int64(l.offset))
		if err != nil && err != io.EOF {
			side_channel.Panic(err)
			return nil
		}
		return []rune(string(buf))
	}
}
