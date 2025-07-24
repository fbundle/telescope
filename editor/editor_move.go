package editor

// moveRelativeAndFixWithoutLock - textCursor is either in the text or at the end of a line
func (e *editor) moveRelativeAndFixWithoutLock(moveRow int, moveCol int) {
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

func (e *editor) MoveLeft() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(0, -1)
		e.setStatusWithoutLock("move left")
	})
}
func (e *editor) MoveRight() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(0, 1)
		e.setStatusWithoutLock("move right")
	})
}
func (e *editor) MoveUp() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(-1, 0)
		e.setStatusWithoutLock("move up")
	})
}
func (e *editor) MoveDown() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(1, 0)
		e.setStatusWithoutLock("move down")
	})
}
func (e *editor) MoveHome() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(0, -e.textCursor.Col)
		e.setStatusWithoutLock("move home")
	})
}
func (e *editor) MoveEnd() {
	e.lockUpdateRender(func() {
		m := e.text
		e.moveRelativeAndFixWithoutLock(0, len(m.Get(e.textCursor.Row))-e.textCursor.Col)
		e.setStatusWithoutLock("move end")
	})
}
func (e *editor) MovePageUp() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(-e.window.height, 0)
		e.setStatusWithoutLock("move page up")
	})
}
func (e *editor) MovePageDown() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(e.window.height, 0)
		e.setStatusWithoutLock("move page down")
	})
}

func (e *editor) Jump(col int, row int) {
	e.lockUpdateRender(func() {
		// move to absolute position
		moveCol := col - e.textCursor.Col
		moveRow := row - e.textCursor.Row
		e.moveRelativeAndFixWithoutLock(moveRow, moveCol)
	})
}
