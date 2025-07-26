package command_editor

import (
	"context"
	"telescope/editor"
	"telescope/log"
	"telescope/text"

	"golang.org/x/exp/mmap"
)

type Cursor struct {
	Row int
	Col int
}

type View struct {
	Mode       string
	WinData    [][]rune
	WinCursor  Cursor
	TextCursor Cursor
	Message    string
	Background string // consider change it to {totalSize, loadedSize, ...}
}

func fromEditorView(view editor.View) View {
	return View{
		Mode:       "",
		WinData:    view.WinData,
		WinCursor:  Cursor{Row: view.WinCursor.Row, Col: view.WinCursor.Col},
		TextCursor: Cursor{Row: view.TextCursor.Row, Col: view.TextCursor.Col},
		Message:    view.Message,
		Background: view.Background,
	}
}

type Controller interface {
	Load(ctx context.Context, inputMmapReader *mmap.ReaderAt) (context.Context, error)
	Resize(height int, width int)

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
	Message(string)

	Escape()
}

type Renderer interface {
	Update() <-chan View
}

type Editor interface {
	Renderer
	Controller
	Text() text.Text
}
