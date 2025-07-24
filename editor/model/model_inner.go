package model

import (
	"io"
	"os"
	"telescope/persistent/vector"

	"golang.org/x/exp/mmap"
)

type lineInFile struct {
	offset int
	size   int
	data   []rune
}

func makeLineFromData(data []rune) lineInFile {
	return lineInFile{
		offset: -1,
		size:   -1,
		data:   data,
	}
}

func makeLineFromFile(offset int, size int) lineInFile {
	return lineInFile{
		offset: offset,
		size:   size,
		data:   nil,
	}
}

func (l lineInFile) repr(r *mmap.ReaderAt) []rune {
	if l.offset < 0 {
		// in-memory lineInFile
		return l.data
	} else {
		buf := make([]byte, l.size)
		_, err := r.ReadAt(buf, int64(l.offset))
		if err != nil && err != io.EOF {
			panic(err)
		}
		return []rune(string(buf))
	}
}
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func EmptyModel() Model {
	return &model{
		r:   nil,
		vec: vector.NewVector[lineInFile](),
	}
}

func LoadModel(filename string, update func(func(Model) Model), done func(), parallel bool) {
	defer done()
	if !fileExists(filename) {
		return
	}
	// model with mmap
	r, err := mmap.Open(filename)
	if err != nil {
		panic(err)
	}

	// TODO - think of any logic to close mmap
	update(func(m Model) Model {
		return &model{
			r:   r,
			vec: m.(*model).vec,
		}
	})

	indexLine := indexLineFile
	if parallel {
		indexLine = indexLineFileParallel
	}

	indexLine(filename, func(offset int, line []byte) {
		update(func(m Model) Model {
			// TODO - potentially update vec every 10000 lines instead of 1
			line := padNewLine(line)
			size := len(line) - endOfLineSize(line)

			vec := m.(*model).vec
			vec = vec.Ins(vec.Len(), makeLineFromFile(offset, size))
			return &model{
				r:   r,
				vec: vec,
			}
		})
	}, done)
}

type model struct {
	r   *mmap.ReaderAt
	vec vector.Vector[lineInFile]
}

func (m *model) Get(i int) []rune {
	return m.vec.Get(i).repr(m.r)
}

func (m *model) Set(i int, val []rune) Model {
	return &model{
		r:   m.r,
		vec: m.vec.Set(i, makeLineFromData(val)),
	}
}

func (m *model) Ins(i int, val []rune) Model {
	return &model{
		r:   m.r,
		vec: m.vec.Ins(i, makeLineFromData(val)),
	}
}

func (m *model) Del(i int) Model {
	return &model{
		r:   m.r,
		vec: m.vec.Del(i),
	}
}

func (m *model) Iter(f func(val []rune) bool) {
	m.vec.Iter(func(l lineInFile) bool {
		return f(l.repr(m.r))
	})
}

func (m *model) Len() int {
	return m.vec.Len()
}
