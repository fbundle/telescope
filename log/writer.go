package log

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

type Writer interface {
	Write(e Entry) Writer
}

func NewWriter(ctx context.Context, filename string) (Writer, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	w := &writer{
		mu:   sync.Mutex{},
		file: f,
	}
	go func() {
		<-ctx.Done()
		w.mu.Lock()
		defer w.mu.Unlock()
		err := w.file.Close()
		if err != nil {
			panic(err)
		}
	}()

	return w, nil
}

type writer struct {
	mu   sync.Mutex
	file *os.File
}

func (w *writer) Write(e Entry) Writer {
	w.mu.Lock()
	defer w.mu.Unlock()

	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	_, err = w.file.Write(append(b, '\n'))
	if err != nil {
		panic(err)
	}
	return w
}

func NewDummyWriter() (Writer, error) {
	return &dummyWriter{}, nil
}

type dummyWriter struct{}

func (w *dummyWriter) Write(Entry) Writer {
	return w
}
