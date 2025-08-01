package text

import (
	"context"
	"telescope/util/bytes"
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

func New(reader bytes.Array) Text {
	return &text{
		reader: reader,
		vec:    vector.New[Line](),
	}
}

func LoadFile(ctx context.Context, reader bytes.Array, update func(Line)) error {
	return indexFile(ctx, reader, '\n', func(offset int, line []byte) {
		line = padNewLine(line)
		size := len(line) - endOfLineSize(line)
		l := makeLineFromFile(offset, size)
		update(l)
	})
}

type text struct {
	reader bytes.Array
	vec    vector.Vector[Line]
}

func (t *text) Append(line Line) Text {
	return &text{
		reader: t.reader,
		vec:    t.vec.Ins(t.vec.Len(), line),
	}
}

func (t *text) Get(i int) []rune {
	return t.vec.Get(i).repr(t.reader)
}

func (t *text) Set(i int, val []rune) Text {
	return &text{
		reader: t.reader,
		vec:    t.vec.Set(i, makeLineFromData(val)),
	}
}

func (t *text) Ins(i int, val []rune) Text {
	return &text{
		reader: t.reader,
		vec:    t.vec.Ins(i, makeLineFromData(val)),
	}
}

func (t *text) Del(i int) Text {
	return &text{
		reader: t.reader,
		vec:    t.vec.Del(i),
	}
}

func (t *text) Iter(f func(i int, val []rune) bool) {
	t.vec.Iter(func(i int, l Line) bool {
		return f(i, l.repr(t.reader))
	})
}

func (t *text) Len() int {
	return t.vec.Len()
}

func (t *text) Split(i int) (Text, Text) {
	v1, v2 := t.vec.Split(i)
	return &text{
			reader: t.reader,
			vec:    v1,
		}, &text{
			reader: t.reader,
			vec:    v2,
		}
}

func (t *text) Concat(t2 Text) Text {
	if t.reader != t2.(*text).reader {
		side_channel.Panic("different readers")
	}
	return &text{
		reader: t.reader,
		vec:    t.vec.Concat(t2.(*text).vec),
	}
}
