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
			CursorRow: uint64(e.textCursor.Row),
			CursorCol: uint64(e.textCursor.Col),
		})

		updateText := func(m text.Text) text.Text {
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
		}

		e.text = updateText(e.text)
		e.moveRelativeAndFixWithoutLock(0, 1) // move right
		e.setStatusWithoutLock("type '%c'", ch)
	})
}

func (e *editor) Backspace() {
	e.lockUpdateRender(func() {
		e.writeLog(log.Entry{
			Command:   log.CommandBackspace,
			CursorRow: uint64(e.textCursor.Row),
			CursorCol: uint64(e.textCursor.Col),
		})

		updateText := func(m text.Text) text.Text {
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
				e.moveRelativeAndFixWithoutLock(-1, len(r1))
			case col != 0:
				newRow := slices.Clone(m.Get(row))
				newRow = deleteFromSlice(newRow, col-1)
				m = m.Set(row, newRow)
				e.moveRelativeAndFixWithoutLock(0, -1)
			default:
				exit.Write("unreachable")
			}
			return m
		}

		e.text = updateText(e.text)
		e.setStatusWithoutLock("backspace")
	})
}

func (e *editor) Delete() {
	e.lockUpdateRender(func() {
		e.writeLog(log.Entry{
			Command:   log.CommandDelete,
			CursorRow: uint64(e.textCursor.Row),
			CursorCol: uint64(e.textCursor.Col),
		})

		updateText := func(m text.Text) text.Text {
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
				exit.Write("unreachable")
			}
			return m
		}

		e.text = updateText(e.text)
		e.setStatusWithoutLock("delete")
	})
}

func (e *editor) Enter() {
	e.lockUpdateRender(func() {
		e.writeLog(log.Entry{
			Command:   log.CommandEnter,
			CursorRow: uint64(e.textCursor.Row),
			CursorCol: uint64(e.textCursor.Col),
		})

		updateText := func(m text.Text) text.Text {
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
				exit.Write("unreachable")
				return m
			}
		}
		e.text = updateText(e.text)
		e.moveRelativeAndFixWithoutLock(1, 0)                 // move down
		e.moveRelativeAndFixWithoutLock(0, -e.textCursor.Col) // move home
		e.setStatusWithoutLock("enter")
	})
}
