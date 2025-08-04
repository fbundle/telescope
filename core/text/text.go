package text

import (
	"telescope/util/buffer"
	seq "telescope/util/persistent/sequence"
	"telescope/util/side_channel"
)

func New(reader buffer.Buffer) Text {
	return Text{
		Reader: reader,
		Lines:  seq.New[Line](),
	}
}

type Text struct {
	Reader buffer.Buffer
	Lines  seq.Seq[Line]
}

func (t Text) Append(line Line) Text {
	return Text{
		Reader: t.Reader,
		Lines:  t.Lines.Ins(t.Lines.Len(), line),
	}
}

func (t Text) Get(i int) []rune {
	return t.Lines.Get(i).Repr(t.Reader)
}

func (t Text) Set(i int, val []rune) Text {
	return Text{
		Reader: t.Reader,
		Lines:  t.Lines.Set(i, MakeLineFromData(val)),
	}
}

func (t Text) Ins(i int, val []rune) Text {
	return Text{
		Reader: t.Reader,
		Lines:  t.Lines.Ins(i, MakeLineFromData(val)),
	}
}

func (t Text) Del(i int) Text {
	return Text{
		Reader: t.Reader,
		Lines:  t.Lines.Del(i),
	}
}

func (t Text) Iter(f func(i int, val []rune) bool) {
	t.Lines.Iter(func(i int, l Line) bool {
		return f(i, l.Repr(t.Reader))
	})
}

func (t Text) Len() int {
	return t.Lines.Len()
}

func (t Text) Split(i int) (Text, Text) {
	v1, v2 := t.Lines.Split(i)
	return Text{
			Reader: t.Reader,
			Lines:  v1,
		}, Text{
			Reader: t.Reader,
			Lines:  v2,
		}
}

func (t Text) Concat(t2 Text) Text {
	if t.Reader != t2.Reader {
		side_channel.Panic("different readers")
	}
	return Text{
		Reader: t.Reader,
		Lines:  t.Lines.Concat(t2.Lines),
	}
}
