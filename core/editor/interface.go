package editor

import (
	"context"
	"telescope/core/log"
	"telescope/core/text"
	"telescope/util/buffer"
)

type Position struct {
	Row int
	Col int
}

type Status struct {
	Message    string
	Background string
	Other      map[string]any
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
	Apply(entry log.Entry)

	InsertLine(lines [][]rune)
	DeleteLine(count int)
}

type Render interface {
	Render() View
	Update() <-chan View
}

type Editor interface {
	Load(ctx context.Context, reader buffer.Buffer) (context.Context, error)
	Escape()
	Resize(height int, width int)
	Status(update func(status Status) Status)
	Render
	Edit
	Move
}
