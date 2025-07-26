package command_editor

import "telescope/editor"

type CommandEditor interface {
	editor.Editor

	Escape()
}
