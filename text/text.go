package text

import (
	"context"
	"telescope/persistent/vector"

	"golang.org/x/exp/mmap"
)

type Text interface {
	Get(i int) []rune
	Set(i int, val []rune) Text
	Ins(i int, val []rune) Text
	Del(i int) Text
	Iter(f func(i int, val []rune) bool)
	Len() int
	Split(i int) (Text, Text)
	Append(line Line) Text
}

func New(r *mmap.ReaderAt) Text {
	return &text{
		reader: r,
		vec:    vector.NewVector[Line](),
	}
}

func LoadFile(ctx context.Context, reader *mmap.ReaderAt, update func(Line)) error {
	return indexFile(ctx, reader, '\n', func(offset int, line []byte) {
		line = padNewLine(line)
		size := len(line) - endOfLineSize(line)
		l := makeLineFromFile(offset, size)
		update(l)
	})
}

type text struct {
	reader *mmap.ReaderAt
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
