package insert_editor

import "telescope/core/editor"

func (e *Editor) gotoAndFixWithoutLock(row int, col int) {
	t := e.text.Get()
	// fix cursor according text
	if t.Len() == 0 {
		row, col = 0, 0
	} else {
		if row < 0 {
			row = 0
		}
		if row >= t.Len() {
			row = t.Len() - 1
		}
		if col < 0 {
			col = 0
		}
		line := t.Get(row)
		if col > len(line) {
			col = len(line) // col can be 1 character outside of text
		}
	}
	// fix window according to cursor
	tlRow, tlCol := e.window.TopLeft.Row, e.window.TopLeft.Col
	width, height := e.window.Dimension.Col, e.window.Dimension.Row
	if row < tlRow {
		tlRow = row
	}
	if row >= tlRow+height {
		tlRow = row - height + 1
	}
	if col < tlCol {
		tlCol = col
	}
	if col >= tlCol+width {
		tlCol = col - width + 1
	}
	// set
	e.cursor = editor.Position{
		Row: row,
		Col: col,
	}
	e.window.TopLeft = editor.Position{
		Row: tlRow,
		Col: tlCol,
	}
}

func (e *Editor) moveRelativeAndFixWithoutLock(moveRow int, moveCol int) {
	e.gotoAndFixWithoutLock(e.cursor.Row+moveRow, e.cursor.Col+moveCol)
}

func (e *Editor) MoveLeft() {
	e.lockRender(func() {
		e.moveRelativeAndFixWithoutLock(0, -1)
		e.setMessageWithoutLock("move left")
	})
}
func (e *Editor) MoveRight() {
	e.lockRender(func() {
		e.moveRelativeAndFixWithoutLock(0, 1)
		e.setMessageWithoutLock("move right")
	})
}
func (e *Editor) MoveUp() {
	e.lockRender(func() {
		e.moveRelativeAndFixWithoutLock(-1, 0)
		e.setMessageWithoutLock("move up")
	})
}
func (e *Editor) MoveDown() {
	e.lockRender(func() {
		e.moveRelativeAndFixWithoutLock(1, 0)
		e.setMessageWithoutLock("move down")
	})
}
func (e *Editor) MoveHome() {
	e.lockRender(func() {
		e.moveRelativeAndFixWithoutLock(0, -e.cursor.Col)
		e.setMessageWithoutLock("move home")
	})
}
func (e *Editor) MoveEnd() {
	e.lockRender(func() {
		t := e.text.Get()
		if e.cursor.Row < t.Len() {
			line := t.Get(e.cursor.Row)
			e.moveRelativeAndFixWithoutLock(0, len(line)-e.cursor.Col)
		}
		e.setMessageWithoutLock("move end")
	})
}
func (e *Editor) MovePageUp() {
	e.lockRender(func() {
		e.moveRelativeAndFixWithoutLock(-e.window.Dimension.Row, 0)
		e.setMessageWithoutLock("move page up")
	})
}
func (e *Editor) MovePageDown() {
	e.lockRender(func() {
		e.moveRelativeAndFixWithoutLock(e.window.Dimension.Row, 0)
		e.setMessageWithoutLock("move page down")
	})
}

func (e *Editor) Goto(row int, col int) {
	e.lockRender(func() {
		e.gotoAndFixWithoutLock(row, col)
		e.setMessageWithoutLock("goto (%d, %d)", row+1, col+1)
	})
}
