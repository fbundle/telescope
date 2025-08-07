package insert_editor

// moveRelativeAndFixWithoutLock - textCursor is either in the text or at the end of a line
func (e *Editor) moveRelativeAndFixWithoutLock(moveRow int, moveCol int) {
	t := e.text.Get()

	e.cursor.Row += moveRow
	e.cursor.Col += moveCol

	// fix text cursor
	if t.Len() == 0 { // NOTE - handle empty file
		e.cursor.Row = 0
		e.cursor.Col = 0
	} else {
		e.cursor.Row = max(0, e.cursor.Row)
		e.cursor.Col = max(0, e.cursor.Col)
		e.cursor.Row = min(e.cursor.Row, t.Len()-1)
		e.cursor.Col = min(e.cursor.Col, len(t.Get(e.cursor.Row))) // textCursor col can be 1 char outside of text
	}

	// fix window
	if e.cursor.Row < e.window.TopLeft.Row {
		e.window.TopLeft.Row = e.cursor.Row
	}
	if e.cursor.Row >= e.window.TopLeft.Row+e.window.Dimension.Row {
		e.window.TopLeft.Row = e.cursor.Row - e.window.Dimension.Row + 1
	}
	if e.cursor.Col < e.window.TopLeft.Col {
		e.window.TopLeft.Col = e.cursor.Col
	}
	if e.cursor.Col >= e.window.TopLeft.Col+e.window.Dimension.Col {
		e.window.TopLeft.Col = e.cursor.Col - e.window.Dimension.Col + 1
	}
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
		// move to absolute position
		moveRow := row - e.cursor.Row
		moveCol := col - e.cursor.Col
		e.moveRelativeAndFixWithoutLock(moveRow, moveCol)
		e.setMessageWithoutLock("goto (%d, %d)", row, col)
	})
}
