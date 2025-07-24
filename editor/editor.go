package editor

import (
	"context"
	"fmt"
	"path/filepath"
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
	inputFilename string
	logWriter     log.Writer
	renderCh      chan View
	reader        *mmap.ReaderAt

	mu         sync.Mutex // the fields below are protected by mu
	loaded     bool
	text       text.Text
	textCursor Cursor
	window     window
	view       internalView
}

func NewEditor(
	ctx context.Context,
	height int,
	width int,
	inputFilename string,
	logFilename string,
	loadDone func(),
) (Editor, error) {

	e := &editor{
		inputFilename: inputFilename,
		logWriter:     nil,
		renderCh:      make(chan View),
		reader:        nil,
		loaded:        false,
		text:          nil,
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
			winName:    "telescope",
			message:    "",
			background: "",
		},
	}

	if len(e.inputFilename) > 0 {
		e.view.winName += " " + filepath.Base(e.inputFilename)
	}
	// closer
	go func() {
		<-ctx.Done()
		e.lockUpdate(func() {
			close(e.renderCh)
			if e.reader != nil {
				_ = e.reader.Close()
			}
		})
	}()

	// log
	var err error
	if len(logFilename) > 0 {
		e.logWriter, err = log.NewWriter(ctx, logFilename)
		if err != nil {
			return nil, err
		}
	} else {
		e.logWriter, err = log.NewDummyWriter()
	}

	// text
	if !fileExists(e.inputFilename) {
		e.reader = nil
		e.text = text.New(nil)
		e.view.message = fmt.Sprintf("file does not exists %s", filepath.Base(e.inputFilename))
		e.loaded = true
		if loadDone != nil {
			loadDone()
		}
	} else {
		r, err := mmap.Open(inputFilename)
		if err != nil {
			return nil, err
		}
		e.reader = r
		e.text = text.New(e.reader)
		// load file asynchronously
		go func() {
			totalSize := fileSize(e.inputFilename)
			loadedSize := 0
			lastPercentage := 0
			t0 := time.Now()
			t1 := t0
			text.LoadFile(ctx, e.inputFilename, func(l text.Line) {
				loadedSize += l.Size()
				e.lockUpdate(func() {
					t2 := time.Now()
					e.text = e.text.Append(l)
					percentage := int(100 * float64(loadedSize) / float64(totalSize))
					if percentage > lastPercentage || t2.Sub(t1) >= feature.LOADING_PROGRESS_INTERVAL_MS*time.Millisecond {
						lastPercentage = percentage
						e.view.background = fmt.Sprintf(
							"loading %s %d/%d (%d%%)",
							filepath.Base(e.inputFilename), loadedSize, totalSize, lastPercentage,
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
					"loaded %s in %d seconds",
					filepath.Base(e.inputFilename), int(totalTime.Seconds()),
				)
				e.loaded = true
			})
			if loadDone != nil {
				loadDone()
			}
		}()
	}

	return e, nil
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
