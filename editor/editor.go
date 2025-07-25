package editor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"telescope/hist"
	"telescope/log"
	"telescope/text"
	"time"

	"golang.org/x/exp/mmap"
)

type internalView struct {
	tlRow      int
	tlCol      int
	height     int
	width      int
	message    string
	background string
}

type editor struct {
	renderCh  chan View // buffered channel is necessary  for preventing deadlock
	logWriter log.Writer

	mu     sync.Mutex // the fields below are protected by mu
	text   hist.Hist[text.Text]
	cursor Cursor
	view   internalView
}

func NewEditor(
	height int, width int,
	logWriter log.Writer,
) (Editor, error) {
	e := &editor{
		renderCh:  make(chan View, 1),
		logWriter: logWriter,

		mu:   sync.Mutex{},
		text: nil,
		cursor: Cursor{
			Row: 0, Col: 0,
		},
		view: internalView{
			tlRow: 0, tlCol: 0,
			width: width, height: height,
			message:    "",
			background: "",
		},
	}
	return e, nil
}

func (e *editor) Load(ctx context.Context, inputMmapReader *mmap.ReaderAt) (context.Context, error) {
	loadCtx, loadDone := context.WithCancel(ctx) // if ctx is done then this editor will also stop loading
	var err error = nil
	e.lockUpdateRender(func() {
		if e.text != nil {
			err = errors.New("load twice")
			return
		}
		e.text = hist.New(text.New(inputMmapReader))
		e.view.background = "loading started"
		go func() { // load file asynchronously
			defer loadDone()
			if inputMmapReader == nil || inputMmapReader.Len() == 0 {
				return // nothing to load
			}

			t0 := time.Now()
			loader := newLoader(inputMmapReader.Len())
			text.LoadFile(ctx, inputMmapReader, func(l text.Line) {
				e.lockUpdate(func() {
					e.text.Update(func(t text.Text) text.Text {
						return t.Append(l)
					})
					if loader.add(l.Size()) { // to render
						e.view.background = fmt.Sprintf(
							"loading %d/%d (%d%%)",
							loader.loadedSize, loader.totalSize, loader.lastRenderPercentage,
						)
						e.renderWithoutLock()
					}
				})
			})
			e.lockUpdateRender(func() {
				totalTime := time.Since(t0)
				e.view.background = ""
				select {
				case <-ctx.Done():
					e.view.message = fmt.Sprintf(
						"loading was cancelled after %d seconds",
						int(totalTime.Seconds()),
					)
				default:
					e.view.message = fmt.Sprintf(
						"loaded for %d seconds",
						int(totalTime.Seconds()),
					)
				}
				e.renderWithoutLock()
			})
		}()
	})

	return loadCtx, err
}

func (e *editor) lockUpdate(f func()) {
	e.mu.Lock()
	defer e.mu.Unlock()

	f()
}

func (e *editor) lockUpdateRender(f func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	defer e.renderWithoutLock()

	f()
}

func (e *editor) Update() <-chan View {
	return e.renderCh
}

func (e *editor) setMessageWithoutLock(format string, a ...any) {
	e.view.message = fmt.Sprintf(format, a...)
}

func (e *editor) writeLog(entry log.Entry) {
	if e.logWriter == nil {
		return
	}
	e.logWriter.Write(entry)
}

func (e *editor) Message(message string) {
	e.lockUpdateRender(func() {
		e.setMessageWithoutLock(message)
	})
}
