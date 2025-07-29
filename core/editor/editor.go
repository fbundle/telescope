package editor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"telescope/config"
	"telescope/core/bytes"
	"telescope/core/hist"
	log2 "telescope/core/log"
	text2 "telescope/core/text"
	"telescope/util/side_channel"
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
	logWriter log2.Writer

	mu         sync.Mutex // the fields below are protected by mu
	text       hist.Hist[text2.Text]
	cursor     Cursor
	windowInfo windowInfo
	status     Status
}

func NewEditor(
	height int, width int,
	logWriter log2.Writer,
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
	e.lockRender(func() {
		if e.text != nil {
			err = errors.New("load twice")
			return
		}
		e.text = hist.New(text2.New(reader))
		e.status.Background = "loading started"
		go func() { // load file asynchronously
			defer loadDone()
			if reader == nil || reader.Len() == 0 {
				return // nothing to load
			}

			t0 := time.Now()
			l := newLoader(reader.Len())
			err = text2.LoadFile(ctx, reader, func(line text2.Line) {
				e.lock(func() {
					e.text.Update(func(t text2.Text) text2.Text {
						return t.Append(line)
					})
					if l.add(line.Size()) { // to makeView
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

func (e *editor) writeLog(entry log2.Entry) {
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
	e.lockRender(func() {
		e.status = update(e.status)
	})
}
