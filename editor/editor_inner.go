package editor

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sync"
	"telescope/editor/model"
	"telescope/flag"
	"time"
)

type Model = model.Model

type status struct {
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
	loadCtx     context.Context
	updateCh    chan struct{}

	mu     sync.Mutex // the fields below are protected by mu
	model  Model
	cursor Cursor
	window window
	status status
}

func NewEditor(height int, width int, filenameIn string, filenameOut string) Editor {
	ctx, cancel := context.WithCancel(context.Background())
	e := &editor{
		filenameIn:  filenameIn,
		filenameOut: filenameOut,
		mu:          sync.Mutex{},
		model:       model.EmptyModel(),
		loadCtx:     ctx,
		cursor: Cursor{
			Row: 0,
			Col: 0,
		},
		window: window{
			tlRow:  0,
			tlCol:  0,
			height: height,
			width:  width,
		},
		status: status{
			loading: "loading 0 lines",
		},
		updateCh: make(chan struct{}, 1),
	}

	parallel := flag.ParallelIndexing()
	t0 := time.Now()
	go model.LoadModel(filenameIn, func(f func(model.Model) model.Model) {
		e.lock(func() {
			e.model = f(e.model)
		})
	}, cancel, parallel)
	go func() {
		defer func() {
			// finally delete loading status
			e.lock(func() {
				elapsed := time.Now().Sub(t0)
				e.status.loading = ""
				e.setStatusWithoutLock("loaded %d lines in %ds", e.model.Len(), int(elapsed.Seconds()))
				if parallel {
					e.setStatusWithoutLock("parallel loaded %d lines in %ds", e.model.Len(), int(elapsed.Seconds()))
				}
			})
			select {
			case e.updateCh <- struct{}{}: //trigger redraw
			default:
			}
		}()
		ticker := time.NewTicker(time.Second) // update loading status every second
		for {
			select {
			case <-ticker.C:
				e.lock(func() {
					elapsed := time.Now().Sub(t0)
					e.status.loading = fmt.Sprintf("loading %d lines in %ds", e.model.Len(), int(elapsed.Seconds()))
					if parallel {
						e.status.loading = fmt.Sprintf("parallel loading %d lines in %ds", e.model.Len(), int(elapsed.Seconds()))
					}
				})
				select {
				case e.updateCh <- struct{}{}: //trigger redraw
				default:
				}
			case <-e.loadCtx.Done():
				return
			}
		}
	}()
	return e
}

func (e *editor) lock(f func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	f()
}

func getRowForView(m Model, row int) []rune {
	if row < m.Len() {
		return m.Get(row)
	} else {
		return []rune{'~'}
	}
}

func (e *editor) Update() <-chan struct{} {
	return e.updateCh
}

func (e *editor) Render() View {
	var m Model
	var win window
	var cur Cursor
	var stat status
	e.lock(func() {
		m = e.model
		win = e.window
		cur = e.cursor
		stat = e.status
	})

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
	view.Status = fmt.Sprintf("(%d, %d)", cur.Row, cur.Col)
	if len(stat.message) > 0 {
		view.Status += " - " + stat.message
	}
	if len(stat.loading) > 0 {
		view.Status += " - " + stat.loading
	}
	return view
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
	e.lock(func() {
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
	e.lock(func() {
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
	e.lock(func() {
		e.moveAndFixWithoutLock(0, -1)
		e.setStatusWithoutLock("move left")
	})
}
func (e *editor) MoveRight() {
	e.lock(func() {
		e.moveAndFixWithoutLock(0, 1)
		e.setStatusWithoutLock("move right")
	})
}
func (e *editor) MoveUp() {
	e.lock(func() {
		e.moveAndFixWithoutLock(-1, 0)
		e.setStatusWithoutLock("move up")
	})
}
func (e *editor) MoveDown() {
	e.lock(func() {
		e.moveAndFixWithoutLock(1, 0)
		e.setStatusWithoutLock("move down")
	})
}
func (e *editor) MoveHome() {
	e.lock(func() {
		e.moveAndFixWithoutLock(0, -e.cursor.Col)
		e.setStatusWithoutLock("move home")
	})
}
func (e *editor) MoveEnd() {
	e.lock(func() {
		m := e.model
		e.moveAndFixWithoutLock(0, len(m.Get(e.cursor.Row))-e.cursor.Col)
		e.setStatusWithoutLock("move end")
	})
}
func (e *editor) MovePageUp() {
	e.lock(func() {
		e.moveAndFixWithoutLock(-e.window.height, 0)
		e.setStatusWithoutLock("move page up")
	})
}
func (e *editor) MovePageDown() {
	e.lock(func() {
		e.moveAndFixWithoutLock(e.window.height, 0)
		e.setStatusWithoutLock("move page down")
	})
}

func (e *editor) Type(ch rune) {
	e.lock(func() {
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
	e.lock(func() {
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
	e.lock(func() {
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
	e.lock(func() {
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
