package editor

type Cursor struct {
	Row int
	Col int
}

type View struct {
	Data   [][]rune // pixels
	Cursor Cursor
	Status []rune
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
	Update() <-chan View
}

type Editor interface {
	Renderer
	Controller
}
