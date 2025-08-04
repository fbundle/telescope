package text

import (
	"telescope/util/buffer"
	"telescope/util/persistent/sequence"
)

type Text2 struct {
	Reader buffer.Buffer
	Lines  sequence.Seq[Line]
}

func New2(reader buffer.Buffer) Text2 {
	return Text2{
		Reader: reader,
		Lines:  sequence.New[Line](),
	}
}
