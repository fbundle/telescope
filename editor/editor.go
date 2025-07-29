package editor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"telescope/bytes"
	"telescope/config"
	"telescope/hist"
	"telescope/log"
	"telescope/text"
	"time"
)

type windowInfo struct {
	tlRow  int
	tlCol  int
	height int
	width  int
}

type editor struct {
	renderCh  chan View
	logWriter log.Writer

	mu         sync.Mutex // the fields below are protected by mu
	text       hist.Hist[text.Text]
	cursor     Cursor
	windowInfo windowInfo
	status     Status
}

func NewEditor(
	height int, width int,
	logWriter log.Writer,
) (Editor, error) {
	e := &editor{
		// buffered channel is necessary  for preventing deadlock
		renderCh:  make(chan View, config.Load().VIEW_CHANNEL_SIZE),
		logWriter: logWriter,

		mu:   sync.Mutex{},
		text: nil,
		cursor: Cursor{
			Row: 0, Col: 0,
		},
		windowInfo: windowInfo{
			tlRow: 0, tlCol: 0,
			height: height, width: width,
		},
		status: Status{
			Header:     "",
			Command:    "",
			Message:    "",
			Background: "",
		},
	}
	return e, nil
}

func (e *editor) Load(ctx context.Context, reader bytes.Array) (context.Context, error) {
	loadCtx, loadDone := context.WithCancel(ctx) // if ctx is done then this editor will also stop loading
	var err error = nil
	e.lockUpdateRender(func() {
		if e.text != nil {
			err = errors.New("load twice")
			return
		}
		e.text = hist.New(text.New(reader))
		e.status.Background = "loading started"
		go func() { // load file asynchronously
			defer loadDone()
			if reader == nil || reader.Len() == 0 {
				return // nothing to load
			}

			t0 := time.Now()
			loader := newLoader(reader.Len())
			text.LoadFile(ctx, reader, func(l text.Line) {
				e.lockUpdate(func() {
					e.text.Update(func(t text.Text) text.Text {
						return t.Append(l)
					})
					if loader.add(l.Size()) { // to renderWithoutLock
						e.status.Background = fmt.Sprintf(
							"loading %d/%d (%d%%)",
							loader.loadedSize, loader.totalSize, loader.lastRenderPercentage,
						)
						e.postWithoutLock()
					}
				})
			})
			e.lockUpdateRender(func() {
				totalTime := time.Since(t0)
				e.status.Background = ""
				select {
				case <-ctx.Done():
					e.status.Message = fmt.Sprintf(
						"loading was cancelled after %d seconds",
						int(totalTime.Seconds()),
					)
				default:
					e.status.Message = fmt.Sprintf(
						"loaded for %d seconds",
						int(totalTime.Seconds()),
					)
				}
				e.postWithoutLock()
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
	defer e.postWithoutLock()

	f()
}

func (e *editor) setMessageWithoutLock(format string, a ...any) {
	e.status.Message = fmt.Sprintf(format, a...)
}

func (e *editor) writeLog(entry log.Entry) {
	if e.logWriter == nil {
		return
	}
	e.logWriter.Write(entry)
}

func (e *editor) WriteMessage(message string) {
	e.lockUpdateRender(func() {
		e.status.Message = message
	})
}
func (e *editor) WriteHeaderCommandMessage(header string, command string, message string) {
	e.lockUpdateRender(func() {
		e.status.Header = header
		e.status.Command = command
		e.status.Message = message
	})
}

func (e *editor) Resize(height int, width int) {
	e.lockUpdateRender(func() {
		if e.windowInfo.height == height && e.windowInfo.width == width {
			return
		}
		e.windowInfo.height, e.windowInfo.width = height, width
		e.moveRelativeAndFixWithoutLock(0, 0)
		e.setMessageWithoutLock("resize to %dx%d", height, width)
	})
}

func (e *editor) Escape() {

}

func (e *editor) Status(update func(status Status) Status) {
	e.lockUpdateRender(func() {
		e.status = update(e.status)
	})
}
