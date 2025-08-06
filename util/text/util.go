package text

import (
	"context"
	"telescope/util/buffer"
	seq "telescope/util/persistent/sequence"
)

const delim byte = '\n'

func LoadFile(ctx context.Context, reader buffer.Reader, update func(Line, int)) error {
	return indexFile(ctx, reader, func(offset int, line []byte) {
		l := makeLineFromOffset(offset)
		update(l, len(line))
	})
}

func indexFile(ctx context.Context, reader buffer.Reader, update func(offset int, line []byte)) error {
	var offset int = 0

	var line []byte

	for i := 0; i < reader.Len(); i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		b := reader.At(i)
		line = append(line, b)
		if b == delim {
			update(offset, line) // line with delim
			offset, line = i+1, nil
		}
	}
	if len(line) > 0 {
		update(offset, line)
	}
	return nil
}

func GetLinesFromSeq(reader buffer.Reader, lines seq.Seq[Line]) [][]rune {
	out := make([][]rune, 0, lines.Len())
	for _, line := range lines.Iter {
		out = append(out, line.Repr(reader))
	}
	return out
}

func GetSeqFromLines(lines [][]rune) seq.Seq[Line] {
	s := seq.Empty[Line]()
	for _, line := range lines {
		s = s.Ins(s.Len(), makeLineFromData(line))
	}
	return s
}
