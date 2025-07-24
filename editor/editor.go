package editor

type Cursor struct {
	Row int
	Col int
}

type View struct {
	Data   [][]rune // pixels
	Cursor Cursor
	Status string
}

type Controller interface {
	Resize(height int, width int)
	Save()

	Type(ch rune)
	Enter()
	Backspace()
	Delete()

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
	Render() View
	Update() <-chan struct{}
}

type Editor interface {
	Renderer
	Controller
}
