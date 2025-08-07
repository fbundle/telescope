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
	cursor editor.Position
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
		cursor: editor.Position{
			Row: 0, Col: 0,
		},
		window: editor.Window{
			TopLeft:   editor.Position{Row: 0, Col: 0},
			Dimension: editor.Position{Row: height, Col: width},
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
		if e.window.Dimension.Row == height && e.window.Dimension.Col == width {
			return
		}
		e.window.Dimension.Row, e.window.Dimension.Col = height, width
		e.moveRelativeAndFixWithoutLock(0, 0)
		e.setMessageWithoutLock("resize to %dx%d", height, width)
	})
}

func (e *Editor) Status(update func(status editor.Status) editor.Status) {
	e.lockRender(func() {
		e.status = update(e.status)
	})
}

func (e *Editor) Load(ctx context.Context, reader buffer.Reader) (context.Context, error) {
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
			err = text.LoadFile(ctx, reader, func(line text.Line, size int) {
				e.lock(func() {
					e.text.Update(func(t text.Text) text.Text {
						return t.AppendLine(line)
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

func (e *Editor) Action(key string, vals ...any) {
	// nothing
}
