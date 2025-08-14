package text

import "telescope/util/buffer"

const delim byte = '\n'

// Line - if offset >= 0, this is a file else this is a []rune buffer
type Line struct {
	Offset int64   // 8 bytes
	Data   *[]rune // 8 bytes on 64-bit system
}

func makeLineFromData(data []rune) Line {
	return Line{
		Offset: -1,
		Data:   &data,
	}
}

func makeLineFromOffset(offset int) Line {
	return Line{
		Offset: int64(offset),
		Data:   nil,
	}
}

func (l Line) Repr(reader buffer.Reader) []rune {
	if l.Offset < 0 {
		// in-memory
		return *l.Data
	} else {
		// from file
		buf := make([]byte, 0)
		for i := int(l.Offset); i < reader.Len(); i++ {
			b := reader.At(i)
			if b == delim {
				break
			}
			buf = append(buf, b)
		}
		return []rune(string(buf))
	}
}
