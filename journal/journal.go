package journal

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"telescope/flag"
	"time"
)

func Read(ctx context.Context, filename string, apply func(entry Entry), done func()) error {
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
		e := Entry{}
		if err := json.Unmarshal(line, &e); err != nil {
			return err
		}
		if flag.Debug() {
			time.Sleep(100 * time.Millisecond)
		}
		apply(e)
		if err == io.EOF {
			return nil
		}
	}
}

type Writer interface {
	Write(e Entry)
}

func NewWriter(ctx context.Context, filename string) (Writer, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	w := &writer{
		f:          f,
		entryQueue: make(chan Entry, 1024),
	}
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C: // flush every 10 seconds
				w.flush()
			}
		}
	}()

	return w, nil
}

type writer struct {
	f          *os.File
	entryQueue chan Entry
	mu         sync.Mutex
}

func (w *writer) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for {
		select {
		case entry := <-w.entryQueue:
			b, err := json.Marshal(entry)
			if err != nil {
				panic(err)
			}
			_, err = w.f.Write(append(b, '\n'))
			if err != nil {
				panic(err)
			}
		default:
			return
		}
	}
}

func (w *writer) Write(e Entry) {
	for {
		select {
		case w.entryQueue <- e:
			return
		default:
			w.flush() // flush if entryQueue is full
		}
	}
}
