package main

import (
	"context"
	"fmt"
	"telescope/util/future"
	"time"
)

func test(ctx context.Context) {
	v := future.New(ctx, func(ctx context.Context) (int, error) {
		s := 0
		for i := 0; i < 100; i++ {
			select {
			case <-ctx.Done():
				return s, ctx.Err()
			default:
			}
			s += i
			time.Sleep(time.Millisecond)
		}
		return s, nil
	})
	<-v.Done()
	if v.Error == nil {
		fmt.Println(v.Value)
	} else {
		panic(v.Error)
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	test(ctx)
	defer cancel()
}
