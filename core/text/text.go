package text

import (
	"telescope/util/buffer"
	"telescope/util/persistent/sequence"
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

func New(reader buffer.Buffer) Text {
	return text{
		Reader: reader,
		Lines:  sequence.New[Line](),
	}
}

type text struct {
	Reader buffer.Buffer
	Lines  sequence.Seq[Line]
}

func (t text) Append(line Line) Text {
	return text{
		Reader: t.Reader,
		Lines:  t.Lines.Ins(t.Lines.Len(), line),
	}
}

func (t text) Get(i int) []rune {
	return t.Lines.Get(i).Repr(t.Reader)
}

func (t text) Set(i int, val []rune) Text {
	return text{
		Reader: t.Reader,
		Lines:  t.Lines.Set(i, makeLineFromData(val)),
	}
}

func (t text) Ins(i int, val []rune) Text {
	return text{
		Reader: t.Reader,
		Lines:  t.Lines.Ins(i, makeLineFromData(val)),
	}
}

func (t text) Del(i int) Text {
	return text{
		Reader: t.Reader,
		Lines:  t.Lines.Del(i),
	}
}

func (t text) Iter(f func(i int, val []rune) bool) {
	t.Lines.Iter(func(i int, l Line) bool {
		return f(i, l.Repr(t.Reader))
	})
}

func (t text) Len() int {
	return t.Lines.Len()
}

func (t text) Split(i int) (Text, Text) {
	v1, v2 := t.Lines.Split(i)
	return text{
			Reader: t.Reader,
			Lines:  v1,
		}, text{
			Reader: t.Reader,
			Lines:  v2,
		}
}

func (t text) Concat(t2 Text) Text {
	if t.Reader != t2.(*text).Reader {
		side_channel.Panic("different readers")
	}
	return text{
		Reader: t.Reader,
		Lines:  t.Lines.Concat(t2.(*text).Lines),
	}
}
