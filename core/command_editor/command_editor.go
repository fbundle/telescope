package command_editor

import (
	"context"
	"sync"
	"telescope/core/editor"
	"telescope/core/log"
	"telescope/core/text"
	"telescope/util/buffer"
	seq "telescope/util/persistent/sequence"
	"telescope/util/side_channel"
)

type Mode = string

const (
	ModeNormal  Mode = "NORMAL"
	ModeCommand Mode = "COMMAND"
	ModeInsert  Mode = "INSERT"
	ModeSelect  Mode = "SELECT"
)

type Selector struct {
	Beg int
	End int
}

type clipboard seq.Seq[text.Line]

type state struct {
	mode      Mode
	command   string
	selector  *Selector
	clipboard clipboard
}

type commandEditor struct {
	cancel            func()
	mu                sync.Mutex
	e                 editor.Editor
	defaultOutputFile string
	state             state
}

func (c *commandEditor) enterNormalModeWithoutLock() {
	c.state.mode = ModeNormal
	c.state.command = ""
	c.state.selector = nil
}
func (c *commandEditor) enterInsertModeWithoutLock() {
	c.state.mode = ModeInsert
	c.state.command = ""
	c.state.selector = nil
}
func (c *commandEditor) enterCommandModeWithoutLock(command string) {
	c.state.mode = ModeCommand
	c.state.command = command
	c.state.selector = nil
}
func (c *commandEditor) enterSelectModeWithoutLock(beg int) {
	c.state.mode = ModeSelect
	c.state.command = ""
	c.state.selector = &Selector{
		Beg: beg,
		End: beg,
	}
}

func (c *commandEditor) Update() <-chan editor.View {
	return c.e.Update()
}

func (c *commandEditor) Load(ctx context.Context, reader buffer.Buffer) (loadCtx context.Context, err error) {
	c.lock(func() {
		loadCtx, err = c.e.Load(ctx, reader)
	})
	return loadCtx, err
}

func (c *commandEditor) Resize(height int, width int) {
	c.lock(func() {
		c.e.Resize(height, width)
	})
}

func (c *commandEditor) Backspace() {
	c.lock(func() {
		switch c.state.mode {
		case ModeNormal:
			// do nothing
		case ModeInsert:
			c.e.Backspace()
		case ModeCommand:
			if len(c.state.command) > 0 {
				c.state.command = c.state.command[:len(c.state.command)-1]
			}
			c.writeWithoutLock("")
		case ModeSelect:
			// do nothing
		default:
			side_channel.Panic("unknown mode: ", c.state)
		}
	})
}

func (c *commandEditor) Delete() {
	c.lock(func() {
		switch c.state.mode {
		case ModeNormal:
			// do nothing
		case ModeInsert:
			c.e.Delete()
		case ModeCommand:
		// do nothing
		case ModeSelect:
			// do nothing
		default:
			side_channel.Panic("unknown mode: ", c.state)
		}
	})
}

func (c *commandEditor) Tabular() {
	c.lock(func() {
		switch c.state.mode {
		case ModeNormal:
			// do nothing
		case ModeInsert:
			c.e.Tabular()
		case ModeCommand:
		// do nothing
		case ModeSelect:
			// do nothing
		default:
			side_channel.Panic("unknown mode: ", c.state)
		}
	})
}

func (c *commandEditor) Goto(row int, col int) {
	c.lock(func() {
		c.e.Goto(row, col)
	})
}

func (c *commandEditor) MoveLeft() {
	c.lock(func() {
		c.e.MoveLeft()
	})
}

func (c *commandEditor) MoveRight() {
	c.lock(func() {
		c.e.MoveRight()
	})
}

func (c *commandEditor) MoveUp() {
	c.lock(func() {
		c.e.MoveUp()
		if c.state.mode == ModeSelect {
			row := c.e.Render().Cursor.Row
			c.state.selector.End = row
			c.writeWithoutLock("select more")
		}
	})
}

func (c *commandEditor) MoveDown() {
	c.lock(func() {
		c.e.MoveDown()
		if c.state.mode == ModeSelect {
			row := c.e.Render().Cursor.Row
			c.state.selector.End = row
			c.writeWithoutLock("select more")
		}
	})
}

func (c *commandEditor) MoveHome() {
	c.lock(func() {
		c.e.MoveHome()
	})
}

func (c *commandEditor) MoveEnd() {
	c.lock(func() {
		c.e.MoveEnd()
	})
}

func (c *commandEditor) MovePageUp() {
	c.lock(func() {
		c.e.MovePageUp()
		if c.state.mode == ModeSelect {
			row := c.e.Render().Cursor.Row
			c.state.selector.End = row
			c.writeWithoutLock("select more")
		}
	})
}

func (c *commandEditor) MovePageDown() {
	c.lock(func() {
		c.e.MovePageDown()
		if c.state.mode == ModeSelect {
			row := c.e.Render().Cursor.Row
			c.state.selector.End = row
			c.writeWithoutLock("select more")
		}
	})
}

func (c *commandEditor) Undo() {
	c.lock(func() {
		c.e.Undo()
	})
}

func (c *commandEditor) Redo() {
	c.lock(func() {
		c.e.Redo()
	})
}

func (c *commandEditor) Apply(entry log.Entry) {
	c.lock(func() {
		c.e.Apply(entry)
	})
}

func (c *commandEditor) Render() editor.View {
	return c.e.Render()
}

func (c *commandEditor) Status(update func(status editor.Status) editor.Status) {
	c.lock(func() {
		c.e.Status(update)
	})
}
func (c *commandEditor) InsertLine(lines seq.Seq[text.Line]) {
	c.lock(func() {
		c.e.InsertLine(lines)
	})
}

func (c *commandEditor) DeleteLine(count int) {
	c.lock(func() {
		c.e.DeleteLine(count)
	})
}

func NewCommandEditor(cancel func(), e editor.Editor, defaultOutputFile string) editor.Editor {
	c := &commandEditor{
		cancel:            cancel,
		mu:                sync.Mutex{},
		e:                 e,
		defaultOutputFile: defaultOutputFile,
		state: state{
			mode:     ModeNormal,
			command:  "",
			selector: nil,
		},
	}
	c.writeWithoutLock("")
	return c
}

func (c *commandEditor) lock(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	f()
}

func (c *commandEditor) writeWithoutLock(message string) {
	c.e.Status(func(status editor.Status) editor.Status {
		if status.Other == nil {
			status.Other = make(map[string]any)
		}
		status.Other["command"] = c.state.command
		status.Other["mode"] = c.state.mode
		status.Other["selector"] = c.state.selector
		status.Message = message
		return status
	})
}
