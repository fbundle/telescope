package buffer

import (
	"runtime"
	"sync"
	"sync/atomic"
	"telescope/util/side_channel"
)

// TODO - experimental for core.text.Line

/*
Chunk contains a single int64
if raw < 0, it is the key to a value of type T in pool
if raw >= 0, it is the byte offset in reader
*/

type Chunk[T any] struct {
	raw int64
}

var pool = &sync.Map{} // map[uint64]any
var lastKey = int64(0)

func NewChunkFromData[T any](data T) *Chunk[T] {
	key := atomic.AddInt64(&lastKey, -1)
	pool.Store(key, data)
	line := &Chunk[T]{
		raw: key,
	}
	runtime.AddCleanup(line, func(line *Chunk[T]) {
		pool.Delete(line.raw)
	}, line)
	return line
}

func NewChunkFromOffset[T any](offset int64) *Chunk[T] {
	if offset < 0 {
		side_channel.Panic("invalid offset")
	}
	return &Chunk[T]{
		raw: offset,
	}
}

func (l *Chunk[T]) Repr(reader Buffer, delim byte, unmarshal func([]byte) T) T {
	i := l.raw
	if i >= 0 {
		buf := make([]byte, 0)
		for i < int64(reader.Len()) {
			b := reader.At(int(i))
			if b == delim {
				break
			}
			buf = append(buf, b)
		}
		return unmarshal(buf)
	} else {
		buf, ok := pool.Load(i)
		if !ok {
			side_channel.Panic("invalid line")
		}
		return buf.(T)
	}
}
