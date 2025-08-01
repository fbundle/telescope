package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Chunk[T any] struct {
	raw int64 // 8 bytes
	_   *bool
}

var pool = &sync.Map{} // map[uint64]any
var lastKey = int64(0)

func NewChunkFromData[T any](data T, cancel func()) *Chunk[T] {
	key := atomic.AddInt64(&lastKey, -1)
	pool.Store(key, data)
	line := &Chunk[T]{
		raw: key,
	}
	runtime.AddCleanup(line, func(key int64) {
		defer cancel()
		pool.Delete(key)
		fmt.Printf("key %d was cleaned\n", key)
	}, key)
	return line
}

func test() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	x := NewChunkFromData[[]byte]([]byte{1, 2, 3}, cancel)
	fmt.Println("size", unsafe.Sizeof(*x))
	_ = x
	return ctx
}

func main() {
	ctx := test()
	runtime.GC()

	fmt.Println("done")

	<-ctx.Done()
}
