package text

import (
	"telescope/util/buffer"
	"telescope/util/persistent/seq"
	"telescope/util/side_channel"
)

func New(reader buffer.Reader) Text {
	return Text{
		Reader: reader,
		Lines:  seq.Empty[Line](),
	}
}

type Text struct {
	Reader buffer.Reader
	Lines  seq.Seq[Line]
}

func (t Text) Get(i int) []rune {
	return t.Lines.Get(i).Repr(t.Reader)
}

func (t Text) Set(i int, val []rune) Text {
	return Text{
		Reader: t.Reader,
		Lines:  t.Lines.Set(i, makeLineFromData(val)),
	}
}

func (t Text) Ins(i int, val []rune) Text {
	return Text{
		Reader: t.Reader,
		Lines:  t.Lines.Ins(i, makeLineFromData(val)),
	}
}

func (t Text) AppendLine(line Line) Text {
	return Text{
		Reader: t.Reader,
		Lines:  t.Lines.Ins(t.Lines.Len(), line),
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

func (t Text) Repr() [][]rune {
	text := make([][]rune, 0, t.Lines.Len())
	for _, line := range t.Lines.Iter {
		text = append(text, line.Repr(t.Reader))
	}
	return text
}

func Slice(t Text, beg int, end int) Text {
	return Text{
		Reader: t.Reader,
		Lines:  seq.Slice(t.Lines, beg, end),
	}
}

func Merge(ts ...Text) Text {
	if len(ts) == 0 {
		side_channel.Panic("cannot merge empty text")
		return Text{}
	}
	t := ts[0]
	for i := 1; i < len(ts); i++ {
		t1 := ts[i]
		reader := t.Reader
		if reader == nil {
			reader = t1.Reader
		} else {
			if t1.Reader != nil && t1.Reader != reader {
				side_channel.Panic("cannot merge text with different reader")
				return Text{}
			}
		}
		t = Text{
			Reader: reader,
			Lines:  seq.Merge(t.Lines, t1.Lines),
		}
	}
	return t
}
