package text

import (
	"telescope/util/buffer"
	"telescope/util/persistent/seq"
)

const delim byte = '\n'

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
