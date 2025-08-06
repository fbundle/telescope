package editor

import (
	"context"
	"telescope/core/text"
	"telescope/util/buffer"
	seq "telescope/util/persistent/sequence"
)

type Position struct {
	Row int
	Col int
}

type Status struct {
	Message    string
	Background string
	Other      map[string]any // arbitrary view
}

type Window struct {
	TopLeft   Position
	Dimension Position
	Data      [][]rune
}

type View struct {
	Text   text.Text
	Cursor Position
	Window Window
	Status Status
}

type Move interface {
	MoveLeft()
	MoveRight()
	MoveUp()
	MoveDown()
	MoveHome()
	MoveEnd()
	MovePageUp()
	MovePageDown()
	Goto(row int, col int)
}

type Edit interface {
	Type(ch rune)
	Backspace()
	Delete()
	Enter()
	Tabular()
	Undo()
	Redo()
	Apply(entry LogEntry)

	InsertLine(lines seq.Seq[text.Line])
	DeleteLine(count int)
}

type Render interface {
	Render() View
	Update() <-chan View
}

type Editor interface {
	Load(ctx context.Context, reader buffer.Reader) (context.Context, error)
	Escape()
	Resize(height int, width int)
	Status(update func(status Status) Status)
	Action(map[string]any) // arbitrary action
	Subscribe(func(LogEntry)) uint64
	Unsubscribe(key uint64)
	Render
	Edit
	Move
}
