package editor

import (
	"context"
	"telescope/bytes"
	"telescope/log"
	"telescope/text"
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
	Data   [][]rune
	Cursor Cursor // relative cursor to window
}

type View struct {
	Window Window
	Status Status
	Cursor Cursor // absolute cursor
	Text   text.Text
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
	Enter()
	Backspace()
	Delete()
	Tabular()
	Undo()
	Redo()
	Apply(entry log.Entry)
}

type App interface {
	WriteHeaderCommandMessage(header string, command string, message string)
	WriteMessage(message string)
}

type Render interface {
	Render() View
	Update() <-chan View
}

type Editor interface {
	Load(ctx context.Context, reader bytes.Array) (context.Context, error)
	Escape()
	Resize(height int, width int)
	Render
	Edit
	Move
	App
}
