package journal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"telescope/flag"
	"time"
)

func Read[T any](ctx context.Context, filename string, apply func(e T), done func()) error {
	defer done()
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		line = bytes.TrimSpace(line)
		if len(line) > 0 {
			var e T
			if err := json.Unmarshal(line, &e); err != nil {
				return err
			}
			if flag.Debug() {
				time.Sleep(100 * time.Millisecond)
			}
			apply(e)
		}
		if err == io.EOF {
			return nil
		}
	}
}

type Writer[T any] interface {
	Write(e T) Writer[T]
}

func NewWriter[T any](ctx context.Context, filename string) (Writer[T], error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	w := &writer[T]{
		mu:      sync.Mutex{},
		file:    f,
		entryCh: make(chan T, 1024),
	}
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				w.flush()
				_ = w.file.Close()
				return
			case <-ticker.C: // flush every 10 seconds
				w.flush()
			}
		}
	}()

	return w, nil
}

type writer[T any] struct {
	mu      sync.Mutex
	file    *os.File
	entryCh chan T
}

func (w *writer[T]) flush() {
	for {
		select {
		case entry := <-w.entryCh:
			b, err := json.Marshal(entry)
			if err != nil {
				panic(err)
			}
			_, err = w.file.Write(append(b, '\n'))
			if err != nil {
				panic(err)
			}
		default:
			return
		}
	}
}

func (w *writer[T]) Write(e T) Writer[T] {
	for {
		select {
		case w.entryCh <- e:
			return w
		default:
			w.flush() // flush if entryCh is full and try again
		}
	}
}
