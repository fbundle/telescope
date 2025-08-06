package text

import (
	"context"
	"telescope/config"
	"telescope/util/buffer"
	seq "telescope/util/persistent/sequence"
)

const delim byte = '\n'

func LoadFile(ctx context.Context, reader buffer.Reader, update func(Line, int)) error {
	memFile := reader.Len() <= config.Load().MAX_MEM_FILE
	return indexFile(ctx, reader, func(offset int, line []byte) {
		var l Line
		if memFile {
			// load directly to memory
			truncLine := line
			if truncLine[len(truncLine)-1] == delim {
				truncLine = truncLine[:len(truncLine)-1]
			}
			l = MakeLineFromData([]rune(string(truncLine)))

		} else {
			l = MakeLineFromOffset(offset)
		}
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
	s := seq.New[Line]()
	for _, line := range lines {
		s = s.Ins(s.Len(), MakeLineFromData(line))
	}
	return s
}
