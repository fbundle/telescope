package text

import (
	"telescope/util/persistent/seq"
)

func MakeTextFromLine(lines [][]rune) Text {
	s := seq.Empty[Line]()
	for _, line := range lines {
		s = s.Ins(s.Len(), makeLineFromData(line))
	}
	return Text{
		reader: nil,
		lines:  s,
	}
}
