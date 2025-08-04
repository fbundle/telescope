package editor

import (
	"slices"
	"telescope/config"
	"telescope/core/log"
	"telescope/core/text"
	seq "telescope/util/persistent/sequence"
	"telescope/util/side_channel"
)

func (e *editor) Type(ch rune) {
	e.lockRender(func() {
		e.writeLog(log.Entry{
			Command: log.CommandType,
			Rune:    ch,
			Row:     uint64(e.cursor.Row),
			Col:     uint64(e.cursor.Col),
		})

		updateText := func(t text.Text) text.Text {
			row, col := e.cursor.Row, e.cursor.Col
			// NOTE - handle empty file
			if t.Len() == 0 {
				t = t.Ins(0, []rune{ch})
				return t
			}
			// t.Get(row) always well-defined
			line := slices.Clone(t.Get(row))
			line = insertToSlice(line, col, ch)
			t = t.Set(row, line)
			return t
		}

		e.text.Update(updateText)
		e.moveRelativeAndFixWithoutLock(0, 1) // move right
		e.setMessageWithoutLock("type '%c'", ch)
	})
}

func (e *editor) Backspace() {
	e.lockRender(func() {
		e.writeLog(log.Entry{
			Command: log.CommandBackspace,
			Row:     uint64(e.cursor.Row),
			Col:     uint64(e.cursor.Col),
		})

		moveRow, moveCol := 0, 0
		updateText := func(t text.Text) text.Text {
			row, col := e.cursor.Row, e.cursor.Col
			// NOTE - handle empty file
			if t.Len() == 0 {
				return t
			}
			switch {
			case col == 0 && row == 0:
			// first line do nothing
			case col == 0 && row != 0:
				// merge 2 lines
				line1 := t.Get(row - 1)
				line2 := t.Get(row)

				t = t.Set(row-1, concatSlices(line1, line2)).Del(row)
				moveRow, moveCol = -1, len(line1) // move up and to the end of last line
			case col != 0:
				// t.Get(row) always well-defined
				line := slices.Clone(t.Get(row))
				line = deleteFromSlice(line, col-1)
				t = t.Set(row, line)
				moveRow, moveCol = 0, -1 // move left
			default:
				side_channel.Panic("unreachable")
			}
			return t
		}
		e.text.Update(updateText)
		e.moveRelativeAndFixWithoutLock(moveRow, moveCol)
		e.setMessageWithoutLock("backspace")
	})
}

func (e *editor) Delete() {
	e.lockRender(func() {
		e.writeLog(log.Entry{
			Command: log.CommandDelete,
			Row:     uint64(e.cursor.Row),
			Col:     uint64(e.cursor.Col),
		})

		updateText := func(t text.Text) text.Text {
			row, col := e.cursor.Row, e.cursor.Col
			// NOTE - handle empty file
			if t.Len() == 0 {
				return t
			}
			// t.Get(row) always well-defined
			line1 := t.Get(row)
			switch {
			case col == len(line1) && row == t.Len()-1:
			// last line, do nothing
			case col == len(line1) && row < t.Len()-1:
				// merge 2 lines
				line2 := t.Get(row + 1)
				t = t.Set(row, concatSlices(line1, line2)).Del(row + 1)
			case col != len(line1):
				line := slices.Clone(line1)
				line = deleteFromSlice(line, col)
				t = t.Set(row, line)
			default:
				side_channel.Panic("unreachable")
			}
			return t
		}
		e.text.Update(updateText)
		e.setMessageWithoutLock("delete")
	})
}

func (e *editor) Enter() {
	e.lockRender(func() {
		e.writeLog(log.Entry{
			Command: log.CommandEnter,
			Row:     uint64(e.cursor.Row),
			Col:     uint64(e.cursor.Col),
		})

		updateText := func(t text.Text) text.Text {
			row, col := e.cursor.Row, e.cursor.Col
			// NOTE - handle empty file
			if t.Len() == 0 {
				t = t.Ins(0, nil)
				return t
			}
			// t.Get(row) always well-defined
			line := t.Get(row)
			switch {
			case col == len(line):
				// add new line
				t = t.Ins(row+1, nil)
				return t
			case col < len(line):
				// split a line
				line1 := slices.Clone(line[:col])
				line2 := slices.Clone(line[col:])
				t = t.Set(row, line1)
				t = t.Ins(row+1, line2)
				return t
			default:
				side_channel.Panic("unreachable")
				return t
			}
		}
		e.text.Update(updateText)
		e.moveRelativeAndFixWithoutLock(1, 0)             // move down
		e.moveRelativeAndFixWithoutLock(0, -e.cursor.Col) // move home
		e.setMessageWithoutLock("enter")
	})
}

func (e *editor) Tabular() {
	for i := 0; i < config.Load().TAB_SIZE; i++ {
		e.Type(' ')
	}
}

func (e *editor) Undo() {
	e.lockRender(func() {
		e.writeLog(log.Entry{
			Command: log.CommandUndo,
		})
		e.text.Undo()
		e.setMessageWithoutLock("undo")
	})
}

func (e *editor) Redo() {
	e.lockRender(func() {
		e.writeLog(log.Entry{
			Command: log.CommandRedo,
		})
		e.text.Redo()
		e.setMessageWithoutLock("redo")
	})
}

func (e *editor) Apply(entry log.Entry) {
	switch entry.Command {
	case log.CommandEnter:
		e.Goto(int(entry.Row), int(entry.Col))
		e.Enter()
	case log.CommandBackspace:
		e.Goto(int(entry.Row), int(entry.Col))
		e.Backspace()
	case log.CommandDelete:
		e.Goto(int(entry.Row), int(entry.Col))
		e.Delete()
	case log.CommandType:
		e.Goto(int(entry.Row), int(entry.Col))
		e.Type(entry.Rune)
	case log.CommandUndo:
		e.Undo()
	case log.CommandRedo:
		e.Redo()
	default:
		side_channel.Panic("command not found")
	}
}

func (e *editor) InsertLine(lines seq.Seq[text.Line]) {
	e.lockRender(func() {
		t := e.text.Get()
		e.writeLog(log.Entry{
			Command: log.CommandInsertLine,
			Text:    text.GetLines(t.Reader, lines),
		})
		row := e.cursor.Row
		update := func(t text.Text) text.Text {
			return text.Text{
				Reader: t.Reader,
				Lines: seq.Concat(
					seq.Slice(t.Lines, 0, row),
					lines,
					seq.Slice(t.Lines, row, t.Lines.Len()),
				),
			}
		}
		e.text.Update(update)
		e.moveRelativeAndFixWithoutLock(lines.Len(), 0)
		e.setMessageWithoutLock("insert lines")
	})
}

func (e *editor) DeleteLine(count int) {
	e.lockRender(func() {
		row := e.cursor.Row
		e.writeLog(log.Entry{
			Command: log.CommandDeleteLine,
			Row:     uint64(row),
			Count:   uint64(count),
		})
		update := func(t text.Text) text.Text {
			return text.Text{
				Reader: t.Reader,
				Lines: seq.Concat(
					seq.Slice(t.Lines, 0, row),
					seq.Slice(t.Lines, row+count, t.Lines.Len()),
				),
			}
		}
		e.text.Update(update)
		e.setMessageWithoutLock("delete lines")
	})
}
