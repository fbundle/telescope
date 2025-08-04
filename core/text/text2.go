package text

import (
	"telescope/util/buffer"
	"telescope/util/persistent/vector"
)

type Text2 struct {
	Reader buffer.Buffer
	Lines  vector.Vector[Line]
}

func New2(reader buffer.Buffer) Text2 {
	return Text2{
		Reader: reader,
		Lines:  vector.New[Line](),
	}
}
