package command_editor

import (
	"context"
	"golang.org/x/exp/mmap"
	"telescope/editor"
	"telescope/exit"
	"telescope/log"
)

// TODO - add INSERT mode and VISUAL mode
// press i VISUAL -> INSERT
// press ESC INSERT -> VISUAL
// add a command buffer, press ESC reset command buffer
type Mode int

const (
	ModeInsert Mode = iota
	ModeCommand
)

type commandEditor struct {
	mode             Mode
	e                editor.Editor
	command          []rune
	latestEditorView editor.View
	renderCh         chan editor.View
}

func (c *commandEditor) Update() <-chan editor.View {
	return c.renderCh
}

func (c *commandEditor) Load(ctx context.Context, inputMmapReader *mmap.ReaderAt) (context.Context, error) {
	return c.e.Load(ctx, inputMmapReader)
}

func (c *commandEditor) Resize(height int, width int) {
	c.e.Resize(height, width)
}

func (c *commandEditor) Type(ch rune) {
	switch c.mode {
	case ModeInsert:
		c.e.Type(ch)
	case ModeCommand:
		c.command = append(c.command, ch)
	default:
		exit.Write("unknown mode: ", c.mode)
	}
}

func (c *commandEditor) Enter() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Backspace() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Delete() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Tabular() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Goto(row int, col int) {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) MoveLeft() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) MoveRight() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) MoveUp() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) MoveDown() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) MoveHome() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) MoveEnd() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) MovePageUp() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) MovePageDown() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Undo() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Redo() {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Apply(entry log.Entry) {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Message(s string) {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Iter(f func(i int, line []rune) bool) {
	//TODO implement me
	panic("implement me")
}

func (c *commandEditor) Escape() {
	//TODO implement me
	panic("implement me")
}

func NewCommandEditor(ctx context.Context, e editor.Editor) CommandEditor {
	c := &commandEditor{
		mode:             ModeCommand,
		e:                e,
		command:          nil,
		latestEditorView: editor.View{},
		renderCh:         make(chan editor.View, 1),
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case view := <-c.e.Update():
				c.latestEditorView = view
			}
		}
	}()
	return c
}
