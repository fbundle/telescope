package editor

import (
	"context"
	"telescope/core/log"
	"telescope/core/text"
	"telescope/util/bytes"
)

type Cursor struct {
	Row int
	Col int
}

type Status struct {
	Header     string
	Command    string
	Message    string
	Background string
}

type Window struct {
	TopLeftRow int
	TopLeftCol int
	Height     int
	Width      int
	Data       [][]rune
}

type View struct {
	Text   text.Text
	Cursor Cursor
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
}

type Render interface {
	Render() View
	Update() <-chan View
}

type Editor interface {
	Load(ctx context.Context, reader bytes.Array) (context.Context, error)
	Escape()
	Resize(height int, width int)
	Status(update func(status Status) Status)
	Render
	Edit
	Move
}
