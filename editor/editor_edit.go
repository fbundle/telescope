package editor

import (
	"slices"
	"telescope/exit"
	"telescope/log"
	"telescope/text"
)

func (e *editor) Type(ch rune) {
	e.lockUpdateRender(func() {
		e.writeLog(log.Entry{
			Command:   log.CommandType,
			Rune:      ch,
			CursorRow: uint64(e.cursor.Row),
			CursorCol: uint64(e.cursor.Col),
		})

		updateText := func(m text.Text) text.Text {
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
		}

		e.text.Update(updateText)
		e.moveRelativeAndFixWithoutLock(0, 1) // move right
		e.setStatusWithoutLock("type '%c'", ch)
	})
}

func (e *editor) Backspace() {
	e.lockUpdateRender(func() {
		e.writeLog(log.Entry{
			Command:   log.CommandBackspace,
			CursorRow: uint64(e.cursor.Row),
			CursorCol: uint64(e.cursor.Col),
		})

		var moveRow, moveCol int
		updateText := func(m text.Text) text.Text {
			row, col := e.cursor.Row, e.cursor.Col
			// NOTE - handle empty file
			if m.Len() == 0 {
				return m
			}
			switch {
			case col == 0 && row == 0:
			// first line do nothing
			case col == 0 && row != 0:
				// merge 2 texts
				r1 := m.Get(row - 1)
				r2 := m.Get(row)

				m = m.Set(row-1, concatSlices(r1, r2)).Del(row)
				moveRow, moveCol = -1, len(r1) // move up and to the end of last line
			case col != 0:
				newRow := slices.Clone(m.Get(row))
				newRow = deleteFromSlice(newRow, col-1)
				m = m.Set(row, newRow)
				moveRow, moveCol = 0, -1 // move left
			default:
				exit.Write("unreachable")
			}
			return m
		}
		e.text.Update(updateText)
		e.moveRelativeAndFixWithoutLock(moveRow, moveCol)
		e.setStatusWithoutLock("backspace")
	})
}

func (e *editor) Delete() {
	e.lockUpdateRender(func() {
		e.writeLog(log.Entry{
			Command:   log.CommandDelete,
			CursorRow: uint64(e.cursor.Row),
			CursorCol: uint64(e.cursor.Col),
		})

		updateText := func(m text.Text) text.Text {
			row, col := e.cursor.Row, e.cursor.Col
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
				exit.Write("unreachable")
			}
			return m
		}
		e.text.Update(updateText)
		e.setStatusWithoutLock("delete")
	})
}

func (e *editor) Enter() {
	e.lockUpdateRender(func() {
		e.writeLog(log.Entry{
			Command:   log.CommandEnter,
			CursorRow: uint64(e.cursor.Row),
			CursorCol: uint64(e.cursor.Col),
		})

		updateText := func(t text.Text) text.Text {
			// NOTE - handle empty file
			if t.Len() == 0 {
				t = t.Ins(0, nil)
				return t
			}
			switch {
			case e.cursor.Col == len(t.Get(e.cursor.Row)):
				// add new line
				t = t.Ins(e.cursor.Row+1, nil)
				return t
			case e.cursor.Col < len(t.Get(e.cursor.Row)):
				// split a line
				r1 := slices.Clone(t.Get(e.cursor.Row)[:e.cursor.Col])
				r2 := slices.Clone(t.Get(e.cursor.Row)[e.cursor.Col:])
				t = t.Set(e.cursor.Row, r1).Ins(e.cursor.Row+1, r2)
				return t
			default:
				exit.Write("unreachable")
				return t
			}
		}
		e.text.Update(updateText)
		e.moveRelativeAndFixWithoutLock(1, 0)             // move down
		e.moveRelativeAndFixWithoutLock(0, -e.cursor.Col) // move home
		e.setStatusWithoutLock("enter")
	})
}

func (e *editor) Undo() {}
func (e *editor) Redo() {}
