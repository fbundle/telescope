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

type View struct {
	Header     string
	WinData    [][]rune
	WinCursor  Cursor
	TextCursor Cursor
	Message    string
	Background string // consider change it to {totalSize, loadedSize, ...}
}

type TextController interface {
	Type(ch rune)
	Enter()
	Backspace()
	Delete()
	Tabular()

	Goto(row int, col int)
	MoveLeft()
	MoveRight()
	MoveUp()
	MoveDown()
	MoveHome()
	MoveEnd()
	MovePageUp()
	MovePageDown()

	Undo()
	Redo()

	Apply(entry log.Entry)
}

type AppController interface {
	Load(ctx context.Context, reader bytes.Array) (context.Context, error)
	Resize(height int, width int)
	WriteHeaderAndMessage(header string, message string)
	WriteMessage(message string)
}

type Renderer interface {
	Cursor() Cursor
	Text() text.Text
	Update() <-chan View
}

type Editor interface {
	Renderer
	AppController
	TextController
}
