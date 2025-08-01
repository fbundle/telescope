package main

import (
	"context"
	"fmt"
	"runtime"
	"telescope/util/buffer"
)

func test() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	x := buffer.NewChunkFromData[[]byte]([]byte{1, 2, 3}, cancel)
	_ = x
	return ctx
}

func main() {
	ctx := test()
	runtime.GC()
	runtime.GC()
	runtime.Gosched()
	fmt.Println("done")

	<-ctx.Done()
}
