package text

import (
	"context"
	"telescope/util/buffer"
	seq "telescope/util/persistent/sequence"
)

const delim byte = '\n'

func LoadFile(ctx context.Context, reader buffer.Buffer, update func(Line, int)) error {
	return indexFile(ctx, reader, func(offset int, line []byte) {
		l := MakeLineFromOffset(offset)
		update(l, len(line))
	})
}

func indexFile(ctx context.Context, reader buffer.Buffer, update func(offset int, line []byte)) error {
	var offset int = 0

	var line []byte

	for i := 0; i < reader.Len(); i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		b := reader.At(i)
		if b != delim {
			line = append(line, b)
			continue
		}
		update(offset, line)
		offset, line = offset+len(line)+1, nil
	}
	if len(line) > 0 {
		update(offset, line)
	}
	return nil
}

func GetLines(reader buffer.Buffer, lines seq.Seq[Line]) [][]rune {
	out := make([][]rune, 0, lines.Len())
	for _, line := range lines.Iter {
		out = append(out, line.Repr(reader))
	}
	return out
}
