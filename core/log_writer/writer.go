package log_writer

import (
	"io"
	"sync"
	"telescope/config"
	"telescope/core/editor"
)

type Writer interface {
	Write(e editor.LogEntry) error
}

func NewWriter(iowriter io.Writer) (Writer, error) {
	// use initial serializer
	version := uint64(config.Load().INITIAL_SERIALIZER_VERSION)
	s, err := GetSerializer(version)
	if err != nil {
		return nil, err
	}
	w := &writer{
		mu:      sync.Mutex{},
		writer:  iowriter,
		marshal: s.Marshal,
	}

	// write set_version using INITIAL_SERIALIZER_VERSION
	// tell reader to use SERIALIZER_VERSION
	err = w.Write(editor.LogEntry{
		Command: editor.CommandSetVersion,
		Version: config.Load().SERIALIZER_VERSION,
	})
	if err != nil {
		return nil, err
	}

	s1, err := GetSerializer(config.Load().SERIALIZER_VERSION)
	if err != nil {
		return nil, err
	}
	w.marshal = s1.Marshal
	return w, nil
}

type writer struct {
	mu      sync.Mutex
	writer  io.Writer
	marshal func(editor.LogEntry) ([]byte, error)
}

func (w *writer) Write(e editor.LogEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	b, err := w.marshal(e)
	if err != nil {
		return err
	}

	return lengthPrefixWrite(w.writer, b)
}
