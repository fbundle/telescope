package editor

import (
	"context"
	"fmt"
	"sync"
	"telescope/feature"
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
	loadCtx   context.Context
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
	ctx context.Context,
	winName string,
	height int, width int,
	inputMmapReader *mmap.ReaderAt,
	logWriter log.Writer,
) (Editor, error) {

	loadCtx, loadCancel := context.WithCancel(ctx) // if ctx is done then this editor will also stop loading

	e := &editor{
		loadCtx:   loadCtx,
		renderCh:  make(chan View),
		logWriter: logWriter,

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

	// text
	e.text = text.New(inputMmapReader)
	if inputMmapReader == nil || inputMmapReader.Len() == 0 {
		loadCancel()
		return e, nil // nothing to load
	}

	// load file asynchronously
	go func() {
		defer loadCancel()

		totalSize := inputMmapReader.Len()
		loadedSize := 0
		lastPercentage := 0
		t0 := time.Now()
		t1 := t0
		text.LoadFile(e.loadCtx, inputMmapReader, func(l text.Line) {
			loadedSize += l.Size()
			e.lockUpdate(func() {
				t2 := time.Now()
				e.text = e.text.Append(l)
				percentage := int(100 * float64(loadedSize) / float64(totalSize))
				if percentage > lastPercentage || t2.Sub(t1) >= feature.LOADING_PROGRESS_INTERVAL_MS*time.Millisecond {
					lastPercentage = percentage
					e.view.background = fmt.Sprintf(
						"loading %d/%d (%d%%)",
						loadedSize, totalSize, lastPercentage,
					)
					t1 = t2
					e.renderWithoutLock()
				}
			})
		})
		e.lockUpdateRender(func() {
			totalTime := time.Now().Sub(t0)
			e.view.background = ""
			e.view.message = fmt.Sprintf(
				"loaded %d seconds",
				int(totalTime.Seconds()),
			)
			e.loaded = true
		})
	}()

	return e, nil
}

func (e *editor) Done() <-chan struct{} {
	return e.loadCtx.Done()
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
