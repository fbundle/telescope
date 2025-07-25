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

	Type(ch rune)
	Enter()
	Backspace()
	Delete()

	Escape()
	Tabular()

	Jump(row int, col int)
	MoveLeft()
	MoveRight()
	MoveUp()
	MoveDown()
	MoveHome()
	MoveEnd()
	MovePageUp()
	MovePageDown()
}

type Renderer interface {
	Update() <-chan View
}

type Text interface {
	Iter(func(i int, line []rune) bool)
}

type Editor interface {
	Renderer
	Controller
	Text
	Done() <-chan struct{}
}
