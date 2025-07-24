package editor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"telescope/editor/model"
	"telescope/flag"
	"time"

	"golang.org/x/exp/mmap"
)

type Model = model.Model

type status struct {
	name    string
	message string
	loading string
}

type window struct {
	tlRow  int
	tlCol  int
	height int
	width  int
}

type editor struct {
	filenameIn  string
	filenameOut string
	renderCh    chan View
	loadCtx     context.Context
	reader      *mmap.ReaderAt

	mu     sync.Mutex // the fields below are protected by mu
	model  Model
	cursor Cursor
	window window
	status status
}

func NewEditor(height int, width int, filenameIn string, filenameOut string) (Editor, error) {
	e := &editor{}

	e.filenameIn = filenameIn
	e.filenameOut = filenameOut

	// some basic properties

	e.cursor = Cursor{
		Row: 0,
		Col: 0,
	}
	e.window = window{
		tlRow:  0,
		tlCol:  0,
		height: height,
		width:  width,
	}
	e.status = status{
		name:    "telescope",
		message: "",
		loading: "",
	}

	if len(e.filenameOut) > 0 {
		e.status.name += " " + filepath.Base(e.filenameOut)
	}

	// model

	e.renderCh = make(chan View, 1)
	var cancelLoadCtx func()
	e.loadCtx, cancelLoadCtx = context.WithCancel(context.Background())
	if !fileExists(e.filenameIn) {
		e.reader = nil
		e.model = model.NewModel(nil)
		e.status.message = fmt.Sprintf("file does not exists %s", filepath.Base(e.filenameIn))
		cancelLoadCtx()
	} else {
		r, err := mmap.Open(filenameIn)
		if err != nil {
			return nil, err
		}
		e.reader = r
		e.model = model.NewModel(e.reader)
		// load file asynchronously
		totalSize := fileSize(e.filenameIn)
		loadedSize := 0
		lastPercentage := 0
		t0 := time.Now()
		t1 := t0
		go model.LoadFile(e.filenameIn, func(l model.Line) {
			loadedSize += l.Size()
			e.lockUpdate(func() {
				t2 := time.Now()
				e.model = e.model.Append(l)
				percentage := int(100 * float64(loadedSize) / float64(totalSize))
				if percentage > lastPercentage || t2.Sub(t1) >= time.Second {
					lastPercentage = percentage
					e.status.loading = fmt.Sprintf("loading %s %d/%d (%d%%)", filepath.Base(e.filenameIn), loadedSize, totalSize, lastPercentage)
					t1 = t2
					e.renderWithoutLock()
				}
			})
		}, func() {
			e.lockUpdateRender(func() {
				totalTime := time.Now().Sub(t0)
				e.status.loading = ""
				e.status.message = fmt.Sprintf("loaded %s in %d seconds", filepath.Base(e.filenameIn), int(totalTime.Seconds()))
			})
			cancelLoadCtx()
		})
	}

	return e, nil
}

func (e *editor) Close() error {
	if e.reader == nil {
		return nil
	}
	return e.reader.Close()
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
	getRowForView := func(m Model, row int) []rune {
		if row < m.Len() {
			return m.Get(row)
		} else {
			return []rune{'~'}
		}
	}
	render := func() View {
		m := e.model
		win := e.window
		cur := e.cursor
		stat := e.status

		view := View{}
		// data
		view.Data = make([][]rune, win.height)
		for row := 0; row < win.height; row++ {
			view.Data[row] = getRowForView(m, row+win.tlRow)
		}
		// cursor
		view.Cursor = Cursor{
			Row: cur.Row - win.tlRow,
			Col: cur.Col - win.tlCol,
		}
		// status
		view.Status = make([]rune, win.width)
		head := []rune(fmt.Sprintf("%s (%d, %d): ", stat.name, cur.Row, cur.Col))
		copy(view.Status, head)

		if len(stat.message) > 0 {
			// write message after head
			copy(view.Status[len(head):], []rune(stat.message))
		}
		if len(stat.loading) > 0 {
			// write loading at the end
			copy(view.Status[len(view.Status)-len(stat.loading):], []rune(stat.loading))
		}

		return view
	}

	e.renderCh <- render()
}

func (e *editor) Update() <-chan View {
	return e.renderCh
}

// moveAndFixWithoutLock - cursor is either in the text or at the end of a line
func (e *editor) moveAndFixWithoutLock(moveRow int, moveCol int) {
	m := e.model

	e.cursor.Row += moveRow
	e.cursor.Col += moveCol

	// fix cursor
	if m.Len() == 0 { // NOTE - handle empty file
		e.cursor.Row = 0
		e.cursor.Col = 0
	} else {
		e.cursor.Row = max(0, e.cursor.Row)
		e.cursor.Col = max(0, e.cursor.Col)
		e.cursor.Row = min(e.cursor.Row, m.Len()-1)
		e.cursor.Col = min(e.cursor.Col, len(m.Get(e.cursor.Row))) // cursor col can be outside of text
	}

	// fix window
	if e.cursor.Row < e.window.tlRow {
		e.window.tlRow = e.cursor.Row
	}
	if e.cursor.Row >= e.window.tlRow+e.window.height {
		e.window.tlRow = e.cursor.Row - e.window.height + 1
	}
	if e.cursor.Col < e.window.tlCol {
		e.window.tlCol = e.cursor.Col
	}
	if e.cursor.Col >= e.window.tlCol+e.window.width {
		e.window.tlCol = e.cursor.Col - e.window.width + 1
	}
}

func (e *editor) setStatusWithoutLock(format string, a ...any) {
	e.status.message = fmt.Sprintf(format, a...)
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
	if len(e.filenameOut) == 0 {
		e.setStatusWithoutLock("read only mode, cannot save")
		return
	}
	select {
	case <-e.loadCtx.Done():
	default:
		e.setStatusWithoutLock("cannot save, still loading")
		return
	}
	// saving is a synchronous task - can be made async but not needed
	var m Model
	e.lockUpdateRender(func() {
		m = e.model
		file, err := os.Create(e.filenameOut)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		for line := range m.Iter {
			if flag.Debug() {
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
		e.moveAndFixWithoutLock(0, -e.cursor.Col)
		e.setStatusWithoutLock("move home")
	})
}
func (e *editor) MoveEnd() {
	e.lockUpdateRender(func() {
		m := e.model
		e.moveAndFixWithoutLock(0, len(m.Get(e.cursor.Row))-e.cursor.Col)
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
		e.model = func(m Model) Model {
			row, col := e.cursor.Row, e.cursor.Col
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
		}(e.model)
		e.moveAndFixWithoutLock(0, 1) // move right
		e.setStatusWithoutLock("type %c", ch)
	})
}

func (e *editor) Backspace() {
	e.lockUpdateRender(func() {
		e.model = func(m Model) Model {
			row, col := e.cursor.Row, e.cursor.Col
			// NOTE - handle empty file
			if m.Len() == 0 {
				return m
			}
			switch {
			case col == 0 && row == 0:
			// first line do nothing
			case col == 0 && row != 0:
				// merge 2 Model
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
		}(e.model)
		e.setStatusWithoutLock("backspace")
	})
}

func (e *editor) Delete() {
	e.lockUpdateRender(func() {
		e.model = func(m Model) Model {
			row, col := e.cursor.Row, e.cursor.Col
			// NOTE - handle empty file
			if m.Len() == 0 {
				return m
			}
			switch {
			case col == len(m.Get(row)) && row == m.Len()-1:
			// last line, do nothing
			case col == len(m.Get(row)) && row < m.Len()-1:
				// merge 2 Model
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
		}(e.model)
		e.setStatusWithoutLock("delete")
	})
}

func (e *editor) Enter() {
	e.lockUpdateRender(func() {
		e.model = func(m Model) Model {
			// NOTE - handle empty file
			if m.Len() == 0 {
				m = m.Ins(0, nil)
				return m
			}
			switch {
			case e.cursor.Col == len(m.Get(e.cursor.Row)):
				// add new line
				m = m.Ins(e.cursor.Row+1, nil)
				return m
			case e.cursor.Col < len(m.Get(e.cursor.Row)):
				// split a line
				r1 := slices.Clone(m.Get(e.cursor.Row)[:e.cursor.Col])
				r2 := slices.Clone(m.Get(e.cursor.Row)[e.cursor.Col:])
				m = m.Set(e.cursor.Row, r1).Ins(e.cursor.Row+1, r2)
				return m
			default:
				panic("unreachable")
			}
		}(e.model)
		e.moveAndFixWithoutLock(1, 0)             // move down
		e.moveAndFixWithoutLock(0, -e.cursor.Col) // move home
		e.setStatusWithoutLock("enter")
	})
}
