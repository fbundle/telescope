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

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	// use initial serializer
	version := uint64(feature.INITIAL_SERIALIZER_VERSION)
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

	// write set version and change to default serializer
	_, err = w.Write(Entry{
		Command: CommandSetVersion,
		Version: feature.DEFAULT_SERIALIZER_VERSION,
	})
	if err != nil {
		return nil, err
	}

	s1, err := getSerializer(feature.DEFAULT_SERIALIZER_VERSION)
	if err != nil {
		return nil, err
	}
	w.marshal = s1.Marshal
	return w, nil
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
