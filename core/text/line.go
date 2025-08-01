package text

import "telescope/util/buffer"

// Line - if offset >= 0, this is a file else this is a []rune buffer
type Line struct {
	offset int64   // 8 bytes
	data   *[]rune // 8 bytes on 64-bit system
}

func makeLineFromData(data []rune) Line {
	return Line{
		offset: -1,
		data:   &data,
	}
}

func makeLineFromFile(offset int) Line {
	return Line{
		offset: int64(offset),
		data:   nil,
	}
}

func (l Line) repr(reader buffer.Buffer) []rune {

	if l.offset < 0 {
		// in-memory
		return *l.data
	} else {
		// from file
		buf := make([]byte, 0)
		for i := int(l.offset); i < reader.Len(); i++ {
			b := reader.At(i)
			if b == '\n' {
				break
			}
			buf = append(buf, b)
		}
		return []rune(string(buf))
	}
}
