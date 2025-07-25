package editor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"telescope/log"
	"telescope/text"
	"time"

	"golang.org/x/exp/mmap"
)

type internalView struct {
	winName    string
	message    string
	background string
}

type window struct {
	tlRow  int
	tlCol  int
	height int
	width  int
}

// TODO - add INSERT mode and VISUAL mode
// press i VISUAL -> INSERT
// press ESC INSERT -> VISUAL
// add a command buffer, press ESC reset command buffer

type editor struct {
	renderCh  chan View
	logWriter log.Writer

	mu         sync.Mutex // the fields below are protected by mu
	loaded     bool
	text       text.Text
	textCursor Cursor
	window     window
	view       internalView
}

func NewEditor(
	winName string,
	height int, width int,
	logWriter log.Writer,
) (Editor, error) {
	e := &editor{
		renderCh:  make(chan View, 1),
		logWriter: logWriter,

		mu:     sync.Mutex{},
		loaded: false,
		text:   nil,
		textCursor: Cursor{
			Row: 0, Col: 0,
		},
		window: window{
			tlRow:  0,
			tlCol:  0,
			height: height,
			width:  width,
		},
		view: internalView{
			winName:    winName,
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
		if e.loaded {
			err = errors.New("load twice")
			return
		}
		e.text = text.New(inputMmapReader)
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
					e.text = e.text.Append(l)
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
				e.view.message = fmt.Sprintf(
					"loaded %d seconds",
					int(totalTime.Seconds()),
				)
				e.loaded = true
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

func (e *editor) setStatusWithoutLock(format string, a ...any) {
	e.view.message = fmt.Sprintf(format, a...)
}

func (e *editor) writeLog(entry log.Entry) {
	if e.logWriter == nil {
		return
	}
	e.logWriter.Write(entry)
}
