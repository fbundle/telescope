package command_editor

import (
	"context"
	"golang.org/x/exp/mmap"
	"sync"
	"telescope/editor"
	"telescope/exit"
	"telescope/log"
	"telescope/text"
)

// TODO - add INSERT mode and VISUAL mode
// press i VISUAL -> INSERT
// press ESC INSERT -> VISUAL
// add a command buffer, press ESC reset command buffer
type Mode = string

const (
	ModeInsert  Mode = "INSERT"
	ModeCommand Mode = "COMMAND"
)

type commandEditor struct {
	mu               sync.Mutex
	mode             Mode
	e                editor.Editor
	command          []rune
	latestEditorView editor.View
	renderCh         chan View
}

func (c *commandEditor) Update() <-chan View {
	return c.renderCh
}

func (c *commandEditor) Load(ctx context.Context, inputMmapReader *mmap.ReaderAt) (loadCtx context.Context, err error) {
	c.lockUpdateRender(func() {
		loadCtx, err = c.e.Load(ctx, inputMmapReader)
	})
	return loadCtx, err
}

func (c *commandEditor) Resize(height int, width int) {
	c.lockUpdateRender(func() {
		c.e.Resize(height, width)
	})
}

func (c *commandEditor) Type(ch rune) {
	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeInsert:
			c.e.Type(ch)
		case ModeCommand:
			c.command = append(c.command, ch)
		default:
			exit.Write("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Enter() {
	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeInsert:
			c.e.Enter()
		case ModeCommand:
			// apply command
			command := string(c.command)
			switch {
			case command == "i":
				c.mode = ModeInsert
				c.command = nil
				c.renderWithoutLock()
			default:
				c.command = nil
				c.e.Message("unknown command: " + command)
				c.renderWithoutLock()
				// TODO - keep writing later - my brain is damn tired
			}
		default:
			exit.Write("unknown mode: ", c.mode)
		}
	})
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

func (c *commandEditor) Text() text.Text {
	return c.e.Text()
}

func (c *commandEditor) renderWithoutLock() {
	view := fromEditorView(c.latestEditorView)
	view.Mode = c.mode
	if len(c.command) > 0 {
		view.Message = string(c.command)
	}
	c.renderCh <- view
}

func (c *commandEditor) lockUpdate(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	f()
}

func (c *commandEditor) lockUpdateRender(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.renderWithoutLock()
}

func NewCommandEditor(ctx context.Context, e editor.Editor) Editor {
	c := &commandEditor{
		mu:               sync.Mutex{},
		mode:             ModeCommand,
		e:                e,
		command:          nil,
		latestEditorView: editor.View{},
		renderCh:         make(chan View, 1),
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case view := <-c.e.Update():
				c.lockUpdateRender(func() {
					c.latestEditorView = view
				})
			}
		}
	}()
	return c
}
