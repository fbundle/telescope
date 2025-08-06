package editor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"telescope/config"
	"telescope/core/hist"
	"telescope/core/log"
	"telescope/core/text"
	"telescope/util/buffer"
	"telescope/util/side_channel"
	"time"
)

type editor struct {
	renderCh  chan View
	logWriter log.Writer

	mu     sync.Mutex // the fields below are protected by mu
	text   hist.Hist[text.Text]
	cursor Position
	window Window
	status Status
}

func New(
	height int, width int,
	logWriter log.Writer,
) (Editor, error) {
	e := &editor{
		// buffered channel is necessary  for preventing deadlock
		renderCh:  make(chan View, config.Load().VIEW_CHANNEL_SIZE),
		logWriter: logWriter,

		mu:   sync.Mutex{},
		text: nil,
		cursor: Position{
			Row: 0, Col: 0,
		},
		window: Window{
			TopLeft:   Position{Row: 0, Col: 0},
			Dimension: Position{Row: height, Col: width},
		},
		status: Status{
			Message:    "",
			Background: "",
			Other:      nil,
		},
	}
	return e, nil
}

func (e *editor) lock(f func()) {
	e.mu.Lock()
	defer e.mu.Unlock()

	f()
}

func (e *editor) lockRender(f func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	defer e.renderWithoutLock()

	f()
}

func (e *editor) setMessageWithoutLock(format string, a ...any) {
	e.status.Message = fmt.Sprintf(format, a...)
}

func (e *editor) writeLog(entry log.Entry) {
	if e.logWriter == nil {
		return
	}
	_, err := e.logWriter.Write(entry)
	if err != nil {
		side_channel.Panic("error write log", err)
		return
	}
}

func (e *editor) Resize(height int, width int) {
	e.lockRender(func() {
		if e.window.Dimension.Row == height && e.window.Dimension.Col == width {
			return
		}
		e.window.Dimension.Row, e.window.Dimension.Col = height, width
		e.moveRelativeAndFixWithoutLock(0, 0)
		e.setMessageWithoutLock("resize to %dx%d", height, width)
	})
}

func (e *editor) Escape() {

}

func (e *editor) Status(update func(status Status) Status) {
	e.lockRender(func() {
		e.status = update(e.status)
	})
}

func (e *editor) Load(ctx context.Context, reader buffer.Buffer) (context.Context, error) {
	loadCtx, loadDone := context.WithCancel(ctx) // if ctx is done then this editor will also stop loading
	var err error = nil
	e.lockRender(func() {
		if e.text != nil {
			err = errors.New("load twice")
			return
		}
		e.text = hist.New(text.New(reader))
		e.status.Background = "loading started"
		go func() { // load file asynchronously
			defer loadDone()
			if reader == nil {
				return // nothing to load
			}

			t0 := time.Now()
			l := newLoader(reader.Len())
			err = text.LoadFile(ctx, reader, func(line text.Line, size int) {
				e.lock(func() {
					e.text.Update(func(t text.Text) text.Text {
						lines := t.Lines
						lines = lines.Ins(lines.Len(), line)
						return text.Text{
							Reader: t.Reader,
							Lines:  lines,
						}
					})
					if l.add(size) { // to makeView
						e.status.Background = fmt.Sprintf(
							"loading %d/%d (%d%%)",
							l.loadedSize, l.totalSize, l.lastRenderPercentage,
						)
						e.renderWithoutLock()
					}
				})
			})
			if err != nil {
				side_channel.Panic("error load file", err)
				return
			}
			e.lockRender(func() {
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
			})
		}()
	})

	return loadCtx, err
}
