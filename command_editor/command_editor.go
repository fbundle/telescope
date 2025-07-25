package command_editor

import "telescope/editor"

func NewCommandEditor(e editor.Editor) CommandEditor {
	return nil
}

// TODO - add INSERT mode and VISUAL mode
// press i VISUAL -> INSERT
// press ESC INSERT -> VISUAL
// add a command buffer, press ESC reset command buffer
type Mode int

const (
	ModeInsert Mode = iota
	ModeVisual
)

type commandEditor struct {
	mode Mode
	e    editor.Editor
}
