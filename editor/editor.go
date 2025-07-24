package editor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"telescope/feature"
	"telescope/journal"
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
	ctx             context.Context
	filenameTextIn  string
	filenameTextOut string
	journalWriter   journal.Writer
	renderCh        chan View
	reader          *mmap.ReaderAt

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
	filenameTextIn string,
	filenameTextOut string,
) (Editor, error) {
	// make journal
	journerWriter, err := journal.NewWriter(ctx, journal.GetJournalFilename(filenameTextIn))
	if err != nil {
		return nil, err
	}

	e := &editor{
		ctx:             ctx,
		filenameTextIn:  filenameTextIn,
		filenameTextOut: filenameTextOut,
		journalWriter:   journerWriter,
		renderCh:        make(chan View),
		reader:          nil,
		loaded:          false,
		text:            nil,
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

	if len(e.filenameTextOut) > 0 {
		e.view.winName += " " + filepath.Base(e.filenameTextOut)
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

	// text
	if !fileExists(e.filenameTextIn) {
		e.reader = nil
		e.text = text.New(nil)
		e.view.message = fmt.Sprintf("file does not exists %s", filepath.Base(e.filenameTextIn))
		e.loaded = true
		// we skip journal file as well
	} else {
		r, err := mmap.Open(filenameTextIn)
		if err != nil {
			return nil, err
		}
		e.reader = r
		e.text = text.New(e.reader)
		// load file asynchronously
		go func() {
			totalSize := fileSize(e.filenameTextIn)
			loadedSize := 0
			lastPercentage := 0
			t0 := time.Now()
			t1 := t0
			text.LoadFile(e.ctx, e.filenameTextIn, func(l text.Line) {
				loadedSize += l.Size()
				e.lockUpdate(func() {
					t2 := time.Now()
					e.text = e.text.Append(l)
					percentage := int(100 * float64(loadedSize) / float64(totalSize))
					if percentage > lastPercentage || t2.Sub(t1) >= time.Second {
						lastPercentage = percentage
						e.view.background = fmt.Sprintf("loading %s %d/%d (%d%%)", filepath.Base(e.filenameTextIn), loadedSize, totalSize, lastPercentage)
						t1 = t2
						e.renderWithoutLock()
					}
				})
			})
			e.lockUpdateRender(func() {
				totalTime := time.Now().Sub(t0)
				e.view.background = ""
				e.view.message = fmt.Sprintf("loaded %s in %d seconds", filepath.Base(e.filenameTextIn), int(totalTime.Seconds()))
				e.loaded = true
			})
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

func (e *editor) renderWithoutLock() {
	getRowForView := func(t text.Text, row int) []rune {
		if row < t.Len() {
			return t.Get(row)
		} else {
			return []rune{'~'}
		}
	}
	render := func() View {

		view := View{
			WinName: e.view.winName,
			WinData: nil,
			WinCursor: Cursor{
				Row: e.textCursor.Row - e.window.tlRow,
				Col: e.textCursor.Col - e.window.tlCol,
			},
			TextCursor: e.textCursor,
			Background: e.view.background,
			Message:    e.view.message,
		}

		// data
		view.WinData = make([][]rune, e.window.height)
		for row := 0; row < e.window.height; row++ {
			view.WinData[row] = getRowForView(e.text, row+e.window.tlRow)
		}
		return view
	}

	e.renderCh <- render()
}

func (e *editor) Update() <-chan View {
	return e.renderCh
}

// moveAndFixWithoutLock - textCursor is either in the text or at the end of a line
func (e *editor) moveAndFixWithoutLock(moveRow int, moveCol int) {
	m := e.text

	e.textCursor.Row += moveRow
	e.textCursor.Col += moveCol

	// fix textCursor
	if m.Len() == 0 { // NOTE - handle empty file
		e.textCursor.Row = 0
		e.textCursor.Col = 0
	} else {
		e.textCursor.Row = max(0, e.textCursor.Row)
		e.textCursor.Col = max(0, e.textCursor.Col)
		e.textCursor.Row = min(e.textCursor.Row, m.Len()-1)
		e.textCursor.Col = min(e.textCursor.Col, len(m.Get(e.textCursor.Row))) // textCursor col can be outside of text
	}

	// fix window
	if e.textCursor.Row < e.window.tlRow {
		e.window.tlRow = e.textCursor.Row
	}
	if e.textCursor.Row >= e.window.tlRow+e.window.height {
		e.window.tlRow = e.textCursor.Row - e.window.height + 1
	}
	if e.textCursor.Col < e.window.tlCol {
		e.window.tlCol = e.textCursor.Col
	}
	if e.textCursor.Col >= e.window.tlCol+e.window.width {
		e.window.tlCol = e.textCursor.Col - e.window.width + 1
	}
}

func (e *editor) setStatusWithoutLock(format string, a ...any) {
	e.view.message = fmt.Sprintf(format, a...)
}

func (e *editor) Resize(height int, width int) {
	e.lockUpdateRender(func() {
		if e.window.height == height && e.window.width == width {
			return
		}
		e.window.height, e.window.width = height, width
		e.moveAndFixWithoutLock(0, 0)
		e.setStatusWithoutLock("resize to %dx%d", height, width)
	})
}

func (e *editor) Save() {
	// saving is a synchronous task - can be made async but not needed
	var m text.Text
	e.lockUpdateRender(func() {
		if len(e.filenameTextOut) == 0 {
			e.setStatusWithoutLock("read only mode, cannot save")
			return
		}
		if !e.loaded {
			e.setStatusWithoutLock("cannot save, still loading")
			return
		}
		m = e.text
		file, err := os.Create(e.filenameTextOut)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		for _, line := range m.Iter {
			if feature.Debug() {
				time.Sleep(100 * time.Millisecond)
			}
			_, err = file.WriteString(string(line) + "\n")
			if err != nil {
				panic(err)
			}
		}
		e.setStatusWithoutLock("saved")
	})
}

func (e *editor) MoveLeft() {
	e.lockUpdateRender(func() {
		e.moveAndFixWithoutLock(0, -1)
		e.setStatusWithoutLock("move left")
	})
}
func (e *editor) MoveRight() {
	e.lockUpdateRender(func() {
		e.moveAndFixWithoutLock(0, 1)
		e.setStatusWithoutLock("move right")
	})
}
func (e *editor) MoveUp() {
	e.lockUpdateRender(func() {
		e.moveAndFixWithoutLock(-1, 0)
		e.setStatusWithoutLock("move up")
	})
}
func (e *editor) MoveDown() {
	e.lockUpdateRender(func() {
		e.moveAndFixWithoutLock(1, 0)
		e.setStatusWithoutLock("move down")
	})
}
func (e *editor) MoveHome() {
	e.lockUpdateRender(func() {
		e.moveAndFixWithoutLock(0, -e.textCursor.Col)
		e.setStatusWithoutLock("move home")
	})
}
func (e *editor) MoveEnd() {
	e.lockUpdateRender(func() {
		m := e.text
		e.moveAndFixWithoutLock(0, len(m.Get(e.textCursor.Row))-e.textCursor.Col)
		e.setStatusWithoutLock("move end")
	})
}
func (e *editor) MovePageUp() {
	e.lockUpdateRender(func() {
		e.moveAndFixWithoutLock(-e.window.height, 0)
		e.setStatusWithoutLock("move page up")
	})
}
func (e *editor) MovePageDown() {
	e.lockUpdateRender(func() {
		e.moveAndFixWithoutLock(e.window.height, 0)
		e.setStatusWithoutLock("move page down")
	})
}

func (e *editor) Type(ch rune) {
	e.lockUpdateRender(func() {
		e.journalWriter.Write(journal.Entry{
			Command:   journal.CommandType,
			Rune:      ch,
			CursorRow: e.textCursor.Row,
			CursorCol: e.textCursor.Col,
		})

		e.text = func(m text.Text) text.Text {
			row, col := e.textCursor.Row, e.textCursor.Col
			// NOTE - handle empty file
			if m.Len() == 0 {
				m = m.Ins(0, []rune{ch})
				return m
			}

			for col >= len(m.Get(row)) {
				newRow := slices.Clone(m.Get(row))
				newRow = append(newRow, ch)
				m = m.Set(row, newRow)
				return m
			}
			newRow := slices.Clone(m.Get(row))
			newRow = insertToSlice(newRow, col, ch)
			m = m.Set(row, newRow)
			return m
		}(e.text)
		e.moveAndFixWithoutLock(0, 1) // move right
		e.setStatusWithoutLock("type '%c'", ch)
	})
}

func (e *editor) Backspace() {
	e.lockUpdateRender(func() {
		e.journalWriter.Write(journal.Entry{
			Command:   journal.CommandBackspace,
			CursorRow: e.textCursor.Row,
			CursorCol: e.textCursor.Col,
		})

		e.text = func(m text.Text) text.Text {
			row, col := e.textCursor.Row, e.textCursor.Col
			// NOTE - handle empty file
			if m.Len() == 0 {
				return m
			}
			switch {
			case col == 0 && row == 0:
			// first line do nothing
			case col == 0 && row != 0:
				// merge 2 Text
				r1 := m.Get(row - 1)
				r2 := m.Get(row)

				m = m.Set(row-1, concatSlices(r1, r2)).Del(row)
				e.moveAndFixWithoutLock(-1, len(r1))
			case col != 0:
				newRow := slices.Clone(m.Get(row))
				newRow = deleteFromSlice(newRow, col-1)
				m = m.Set(row, newRow)
				e.moveAndFixWithoutLock(0, -1)
			default:
				panic("unreachable")
			}
			return m
		}(e.text)
		e.setStatusWithoutLock("backspace")
	})
}

func (e *editor) Delete() {
	e.lockUpdateRender(func() {
		e.journalWriter.Write(journal.Entry{
			Command:   journal.CommandDelete,
			CursorRow: e.textCursor.Row,
			CursorCol: e.textCursor.Col,
		})

		e.text = func(m text.Text) text.Text {
			row, col := e.textCursor.Row, e.textCursor.Col
			// NOTE - handle empty file
			if m.Len() == 0 {
				return m
			}
			switch {
			case col == len(m.Get(row)) && row == m.Len()-1:
			// last line, do nothing
			case col == len(m.Get(row)) && row < m.Len()-1:
				// merge 2 Text
				r1 := m.Get(row)
				r2 := m.Get(row + 1)
				m = m.Set(row, concatSlices(r1, r2)).Del(row + 1)
			case col != len(m.Get(row)):
				newRow := slices.Clone(m.Get(row))
				newRow = deleteFromSlice(newRow, col)
				m = m.Set(row, newRow)
			default:
				panic("unreachable")
			}
			return m
		}(e.text)
		e.setStatusWithoutLock("delete")
	})
}

func (e *editor) Enter() {
	e.lockUpdateRender(func() {
		e.journalWriter.Write(journal.Entry{
			Command:   journal.CommandEnter,
			CursorRow: e.textCursor.Row,
			CursorCol: e.textCursor.Col,
		})

		e.text = func(m text.Text) text.Text {
			// NOTE - handle empty file
			if m.Len() == 0 {
				m = m.Ins(0, nil)
				return m
			}
			switch {
			case e.textCursor.Col == len(m.Get(e.textCursor.Row)):
				// add new line
				m = m.Ins(e.textCursor.Row+1, nil)
				return m
			case e.textCursor.Col < len(m.Get(e.textCursor.Row)):
				// split a line
				r1 := slices.Clone(m.Get(e.textCursor.Row)[:e.textCursor.Col])
				r2 := slices.Clone(m.Get(e.textCursor.Row)[e.textCursor.Col:])
				m = m.Set(e.textCursor.Row, r1).Ins(e.textCursor.Row+1, r2)
				return m
			default:
				panic("unreachable")
			}
		}(e.text)
		e.moveAndFixWithoutLock(1, 0)                 // move down
		e.moveAndFixWithoutLock(0, -e.textCursor.Col) // move home
		e.setStatusWithoutLock("enter")
	})
}

func (e *editor) Escape() {
	// do nothing
}
func (e *editor) Tabular() {
	// tab is two spaces
	e.Type(' ')
	e.Type(' ')
}
