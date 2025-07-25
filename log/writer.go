package log

import (
	"context"
	"os"
	"sync"
	"telescope/feature"
)

type Writer interface {
	Write(e Entry) (Writer, error)
}

func NewWriter(ctx context.Context, filename string) (Writer, error) {
	version := feature.SERIALIZER_VERSION

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	s, err := getSerializer(version)
	if err != nil {
		return nil, err
	}
	w := &writer{
		mu:      sync.Mutex{},
		file:    f,
		marshal: s.Marshal,
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

	// write set version
	return w.Write(Entry{
		Command: CommandSetVersion,
		Version: version,
	})
}

type writer struct {
	mu      sync.Mutex
	file    *os.File
	marshal func(Entry) ([]byte, error)
}

func (w *writer) Write(e Entry) (Writer, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	b, err := w.marshal(e)
	if err != nil {
		return nil, err
	}
	_, err = w.file.Write(append(b, '\n'))
	if err != nil {
		return nil, err
	}
	return w, nil
}
