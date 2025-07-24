package text

import (
	"telescope/flag"
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
	Append(line Line) Text
}

func New(r *mmap.ReaderAt) Text {
	return &model{
		r:   r,
		vec: vector.NewVector[Line](),
	}
}

func LoadFile(filename string, update func(Line), done func()) {
	indexFileFunc := indexFile
	if flag.ParallelIndexing() {
		indexFileFunc = indexFileParallel
	}
	indexFileFunc(filename, func(offset int, line []byte) {
		line = padNewLine(line)
		size := len(line) - endOfLineSize(line)
		l := makeLineFromFile(offset, size)
		update(l)
	}, done)
}

type model struct {
	r   *mmap.ReaderAt
	vec vector.Vector[Line]
}

func (m *model) Append(line Line) Text {
	return &model{
		r:   m.r,
		vec: m.vec.Ins(m.vec.Len(), line),
	}
}

func (m *model) Get(i int) []rune {
	return m.vec.Get(i).repr(m.r)
}

func (m *model) Set(i int, val []rune) Text {
	return &model{
		r:   m.r,
		vec: m.vec.Set(i, makeLineFromData(val)),
	}
}

func (m *model) Ins(i int, val []rune) Text {
	return &model{
		r:   m.r,
		vec: m.vec.Ins(i, makeLineFromData(val)),
	}
}

func (m *model) Del(i int) Text {
	return &model{
		r:   m.r,
		vec: m.vec.Del(i),
	}
}

func (m *model) Iter(f func(i int, val []rune) bool) {
	m.vec.Iter(func(i int, l Line) bool {
		return f(i, l.repr(m.r))
	})
}

func (m *model) Len() int {
	return m.vec.Len()
}
