package journal

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"telescope/feature"
	"time"
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
		mu:      sync.Mutex{},
		file:    f,
		entryCh: make(chan Entry, 1024),
	}
	go func() {
		ticker := time.NewTicker(feature.JOURNAL_INTERVAL_S * time.Second)
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

type writer struct {
	mu      sync.Mutex
	file    *os.File
	entryCh chan Entry
}

func (w *writer) flush() {
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

func (w *writer) Write(e Entry) Writer {
	for {
		select {
		case w.entryCh <- e:
			return w
		default:
			w.flush() // flush if entryCh is full and try again
		}
	}
}

func NewDummyWriter() (Writer, error) {
	return &dummyWriter{}, nil
}

type dummyWriter struct{}

func (w *dummyWriter) Write(Entry) Writer {
	return w
}
