package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/fbundle/go_util/pkg/unbounded"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ch := unbounded.New[int](ctx)

	go func() {
		for i := 0; i < 1_000_000_000; i++ {
			ch.InChan() <- i
		}
		close(ch.InChan())
	}()

	for i := range ch.OutChan() {
		_ = i
	}

}
