package editor

type Cursor struct {
	Row int
	Col int
}

type View struct {
	WinName    string
	WinData    [][]rune
	WinCursor  Cursor
	TextCursor Cursor
	Background string
	Message    string
}

type Controller interface {
	Resize(height int, width int)
	Save()

	Type(ch rune)
	Enter()
	Backspace()
	Delete()

	Escape()
	Tabular()

	MoveLeft()
	MoveRight()
	MoveUp()
	MoveDown()
	MoveHome()
	MoveEnd()
	MovePageUp()
	MovePageDown()
	Jump(col int, row int)
}

type Renderer interface {
	Update() <-chan View
}

type Editor interface {
	Renderer
	Controller
}
