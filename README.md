# TELESCOPE

an extremely fast text editor

![screenshot](./screenshots/0_1_2.png)

## FEATURE SET

here are the projected feature sets

- `nano`-like basic text editor

- extremely fast start up, extremely fast edit

- able to recover from crash

- able to handle very large files, potentially even larger than system memory using persistent vector

## EDITOR INTERFACE

```go
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
}

type Renderer interface {
	Update() <-chan View
}

type Editor interface {
	Renderer
	Controller
}

```

## TECH

- [gocui](https://github.com/jroimartin/gocui)

- persistent vector based on weight-balanced tree

## NOTE
