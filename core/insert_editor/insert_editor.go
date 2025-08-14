package insert_editor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"telescope/config"
	"telescope/core/editor"
	"telescope/util/buffer"
	"telescope/util/hist"
	"telescope/util/side_channel"
	"telescope/util/subsciber_pool"
	"telescope/util/text"
	"time"
)

type Editor struct {
	renderCh chan editor.View

	mu     sync.Mutex // the fields below are protected by mu
	text   *hist.Hist[text.Text]
	cursor editor.Cursor
	window editor.Window
	status editor.Status
	pool   *subsciber_pool.Pool[func(editor.LogEntry)]
}

func New(
	height int, width int,
) (*Editor, error) {
	e := &Editor{
		// buffered channel is necessary  for preventing deadlock
		renderCh: make(chan editor.View, config.Load().VIEW_CHANNEL_SIZE),

		mu:   sync.Mutex{},
		text: nil, // awaiting Load
		cursor: editor.Cursor{
			Row: 0, Col: 0,
		},
		window: editor.Window{
			TlRow:  0,
			TlCol:  0,
			Width:  width,
			Height: height,
		},
		status: editor.Status{
			Message:    "",
			Background: "",
			Other:      nil,
		},
		pool: subsciber_pool.New[func(editor.LogEntry)](),
	}
	return e, nil
}

func (e *Editor) lock(f func()) {
	e.mu.Lock()
	defer e.mu.Unlock()

	f()
}

func (e *Editor) lockRender(f func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	defer e.renderWithoutLock()

	f()
}

func (e *Editor) setMessageWithoutLock(format string, a ...any) {
	e.status.Message = fmt.Sprintf(format, a...)
}

func (e *Editor) writeLogWithoutLock(entry editor.LogEntry) {
	go func() {
		for _, consume := range e.pool.Iter {
			consume(entry)
		}
	}()
}

func (e *Editor) Subscribe(consume func(editor.LogEntry)) uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.pool.Subscribe(consume)
}

func (e *Editor) Unsubscribe(key uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pool.Unsubscribe(key)
}

func (e *Editor) Resize(height int, width int) {
	e.lockRender(func() {
		if e.window.Height == height && e.window.Width == width {
			return
		}
		e.window.Height, e.window.Width = height, width
		e.moveRelativeAndFixWithoutLock(0, 0)
		e.setMessageWithoutLock("resize to %dx%d", height, width)
	})
}

func (e *Editor) Status(update func(status editor.Status) editor.Status) {
	e.lockRender(func() {
		e.status = update(e.status)
	})
}

func (e *Editor) Load(ctx context.Context, reader buffer.Reader, lines ...text.Line) (context.Context, error) {
	loadCtx, loadDone := context.WithCancel(context.Background())
	var err error = nil
	e.lockRender(func() {
		if e.text != nil {
			loadDone()
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
			updateWithoutLock := func(line text.Line) {
				e.text.Update(func(t text.Text) text.Text {
					return t.AppendLine(line)
				})
				if l.set(int(line.Offset())) { // to makeView
					e.status.Background = fmt.Sprintf(
						"loading %d/%d (%d%%)",
						l.loadedSize, l.totalSize, l.lastRenderPercentage,
					)
					e.renderWithoutLock()
				}
			}
			e.lock(func() {
				for _, line := range lines {
					updateWithoutLock(line)
				}
			})

			if len(lines) > 0 {
				err = text.LoadFileAfter(ctx, reader, func(line text.Line) {
					e.lock(func() {
						updateWithoutLock(line)
					})
				}, lines[len(lines)-1])

			} else {
				err = text.LoadFile(ctx, reader, func(line text.Line) {
					e.lock(func() {
						updateWithoutLock(line)
					})
				})
			}
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

func (e *Editor) Action(key string, vals ...any) {
	// nothing
}
