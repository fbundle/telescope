package editor

// moveRelativeAndFixWithoutLock - textCursor is either in the text or at the end of a line
func (e *editor) moveRelativeAndFixWithoutLock(moveRow int, moveCol int) {
	t := e.text.Get()

	e.cursor.Row += moveRow
	e.cursor.Col += moveCol

	// fix textCursor
	if t.Len() == 0 { // NOTE - handle empty file
		e.cursor.Row = 0
		e.cursor.Col = 0
	} else {
		e.cursor.Row = max(0, e.cursor.Row)
		e.cursor.Col = max(0, e.cursor.Col)
		e.cursor.Row = min(e.cursor.Row, t.Len()-1)
		e.cursor.Col = min(e.cursor.Col, len(t.Get(e.cursor.Row))) // textCursor col can be outside of text
	}

	// fix window
	if e.cursor.Row < e.view.tlRow {
		e.view.tlRow = e.cursor.Row
	}
	if e.cursor.Row >= e.view.tlRow+e.view.height {
		e.view.tlRow = e.cursor.Row - e.view.height + 1
	}
	if e.cursor.Col < e.view.tlCol {
		e.view.tlCol = e.cursor.Col
	}
	if e.cursor.Col >= e.view.tlCol+e.view.width {
		e.view.tlCol = e.cursor.Col - e.view.width + 1
	}
}

func (e *editor) MoveLeft() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(0, -1)
		e.setMessageWithoutLock("move left")
	})
}
func (e *editor) MoveRight() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(0, 1)
		e.setMessageWithoutLock("move right")
	})
}
func (e *editor) MoveUp() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(-1, 0)
		e.setMessageWithoutLock("move up")
	})
}
func (e *editor) MoveDown() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(1, 0)
		e.setMessageWithoutLock("move down")
	})
}
func (e *editor) MoveHome() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(0, -e.cursor.Col)
		e.setMessageWithoutLock("move home")
	})
}
func (e *editor) MoveEnd() {
	e.lockUpdateRender(func() {
		t := e.text.Get()
		e.moveRelativeAndFixWithoutLock(0, len(t.Get(e.cursor.Row))-e.cursor.Col)
		e.setMessageWithoutLock("move end")
	})
}
func (e *editor) MovePageUp() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(-e.view.height, 0)
		e.setMessageWithoutLock("move page up")
	})
}
func (e *editor) MovePageDown() {
	e.lockUpdateRender(func() {
		e.moveRelativeAndFixWithoutLock(e.view.height, 0)
		e.setMessageWithoutLock("move page down")
	})
}

func (e *editor) Goto(row int, col int) {
	e.lockUpdateRender(func() {
		// move to absolute position
		moveRow := row - e.cursor.Row
		moveCol := col - e.cursor.Col
		e.moveRelativeAndFixWithoutLock(moveRow, moveCol)
	})
}
