package text

import (
	"telescope/util/buffer"
	"telescope/util/persistent/vector"
	"telescope/util/side_channel"
)

type Text interface {
	Get(i int) []rune
	Set(i int, val []rune) Text
	Ins(i int, val []rune) Text
	Del(i int) Text
	Iter(f func(i int, val []rune) bool)
	Len() int
	Split(i int) (Text, Text)
	Concat(t2 Text) Text
	Append(line Line) Text
}

func New(reader buffer.Buffer) BufferedText {
	return BufferedText{
		reader: reader,
		vec:    vector.New[Line](),
	}
}

type BufferedText struct {
	reader buffer.Buffer
	vec    vector.Vector[Line]
}

func (t BufferedText) Append(line Line) Text {
	return BufferedText{
		reader: t.reader,
		vec:    t.vec.Ins(t.vec.Len(), line),
	}
}

func (t BufferedText) Get(i int) []rune {
	return t.vec.Get(i).repr(t.reader)
}

func (t BufferedText) Set(i int, val []rune) Text {
	return BufferedText{
		reader: t.reader,
		vec:    t.vec.Set(i, makeLineFromData(val)),
	}
}

func (t BufferedText) Ins(i int, val []rune) Text {
	return BufferedText{
		reader: t.reader,
		vec:    t.vec.Ins(i, makeLineFromData(val)),
	}
}

func (t BufferedText) Del(i int) Text {
	return BufferedText{
		reader: t.reader,
		vec:    t.vec.Del(i),
	}
}

func (t BufferedText) Iter(f func(i int, val []rune) bool) {
	t.vec.Iter(func(i int, l Line) bool {
		return f(i, l.repr(t.reader))
	})
}

func (t BufferedText) Len() int {
	return t.vec.Len()
}

func (t BufferedText) Split(i int) (Text, Text) {
	v1, v2 := t.vec.Split(i)
	return BufferedText{
			reader: t.reader,
			vec:    v1,
		}, BufferedText{
			reader: t.reader,
			vec:    v2,
		}
}

func (t BufferedText) Concat(t2 Text) Text {
	if t.reader != t2.(*BufferedText).reader {
		side_channel.Panic("different readers")
	}
	return BufferedText{
		reader: t.reader,
		vec:    t.vec.Concat(t2.(*BufferedText).vec),
	}
}
