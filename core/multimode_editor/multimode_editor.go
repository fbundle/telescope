package multimode_editor

import (
	"context"
	"sync"
	"telescope/config"
	"telescope/core/editor"
	"telescope/core/insert_editor"
	"telescope/util/buffer"
	seq "telescope/util/persistent/sequence"
	"telescope/util/side_channel"
	"telescope/util/text"
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

type clipboard = seq.Seq[text.Line]

type state struct {
	mode      Mode
	command   string
	selector  *Selector
	clipboard clipboard
}

type Editor struct {
	cancel            func()
	mu                sync.Mutex
	e                 *insert_editor.Editor
	defaultOutputFile string
	state             state
}

func (c *Editor) enterNormalModeWithoutLock() {
	c.state.mode = ModeNormal
	c.state.command = ""
	c.state.selector = nil
}
func (c *Editor) enterInsertModeWithoutLock() {
	c.state.mode = ModeInsert
	c.state.command = ""
	c.state.selector = nil
}
func (c *Editor) enterCommandModeWithoutLock(command string) {
	c.state.mode = ModeCommand
	c.state.command = command
	c.state.selector = nil
}
func (c *Editor) enterSelectModeWithoutLock(beg int) {
	c.state.mode = ModeSelect
	c.state.command = ""
	c.state.selector = &Selector{
		Beg: beg,
		End: beg,
	}
}

func (c *Editor) Update() <-chan editor.View {
	return c.e.Update()
}

func (c *Editor) Load(ctx context.Context, reader buffer.Reader) (loadCtx context.Context, err error) {
	c.lock(func() {
		loadCtx, err = c.e.Load(ctx, reader)
	})
	return loadCtx, err
}

func (c *Editor) Resize(height int, width int) {
	c.lock(func() {
		c.e.Resize(height, width)
	})
}

func (c *Editor) Backspace() {
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

func (c *Editor) Delete() {
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

func (c *Editor) Tabular() {
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

func (c *Editor) Goto(row int, col int) {
	c.lock(func() {
		c.e.Goto(row, col)
	})
}

func (c *Editor) MoveLeft() {
	c.lock(func() {
		c.e.MoveLeft()
	})
}

func (c *Editor) MoveRight() {
	c.lock(func() {
		c.e.MoveRight()
	})
}

func (c *Editor) MoveUp() {
	c.lock(func() {
		c.e.MoveUp()
		if c.state.mode == ModeSelect {
			row := c.e.Render().Cursor.Row
			c.state.selector.End = row
			c.writeWithoutLock("select more")
		}
	})
}

func (c *Editor) MoveDown() {
	c.lock(func() {
		c.e.MoveDown()
		if c.state.mode == ModeSelect {
			row := c.e.Render().Cursor.Row
			c.state.selector.End = row
			c.writeWithoutLock("select more")
		}
	})
}

func (c *Editor) MoveHome() {
	c.lock(func() {
		c.e.MoveHome()
	})
}

func (c *Editor) MoveEnd() {
	c.lock(func() {
		c.e.MoveEnd()
	})
}

func (c *Editor) MovePageUp() {
	c.lock(func() {
		c.e.MovePageUp()
		c.maybeUpdateSelectorEndWithoutLock()
	})
}

func (c *Editor) MovePageDown() {
	c.lock(func() {
		c.e.MovePageDown()
		c.maybeUpdateSelectorEndWithoutLock()
	})
}

func (c *Editor) Undo() {
	c.lock(func() {
		c.e.Undo()
	})
}

func (c *Editor) Redo() {
	c.lock(func() {
		c.e.Redo()
	})
}

func (c *Editor) Apply(entry editor.LogEntry) {
	c.lock(func() {
		c.e.Apply(entry)
	})
}

func (c *Editor) Render() editor.View {
	return c.e.Render()
}

func (c *Editor) Status(update func(status editor.Status) editor.Status) {
	c.lock(func() {
		c.e.Status(update)
	})
}
func (c *Editor) InsertLine(lines seq.Seq[text.Line]) {
	c.lock(func() {
		c.e.InsertLine(lines)
	})
}

func (c *Editor) DeleteLine(count int) {
	c.lock(func() {
		c.e.DeleteLine(count)
	})
}

func New(cancel func(), e *insert_editor.Editor, defaultOutputFile string) *Editor {
	c := &Editor{
		cancel:            cancel,
		mu:                sync.Mutex{},
		e:                 e,
		defaultOutputFile: defaultOutputFile,
		state: state{
			mode:      ModeNormal,
			command:   "",
			selector:  nil,
			clipboard: seq.Empty[text.Line](),
		},
	}
	c.writeWithoutLock("")
	return c
}

func (c *Editor) lock(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	f()
}

func (c *Editor) writeWithoutLock(message string) {
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

func (c *Editor) Action(action map[string]any) {
	if action == nil {
		return
	}
	if o, ok := action["mouse_click"]; ok {
		p := o.(editor.Position)
		row, col := p.Row, p.Col
		tl := c.e.Render().Window.TopLeft
		c.e.Goto(tl.Row+row, tl.Col+col)
	}
	if _, ok := action["mouse_scroll_up"]; ok {
		for i := 0; i < config.Load().SCROLL_SPEED; i++ {
			c.e.MoveUp()
		}
	}
	if _, ok := action["mouse_scroll_down"]; ok {
		for i := 0; i < config.Load().SCROLL_SPEED; i++ {
			c.e.MoveDown()
		}
	}
}

func (c *Editor) Subscribe(consume func(editor.LogEntry)) uint64 {
	var key uint64
	c.lock(func() {
		key = c.e.Subscribe(consume)
	})
	return key
}
func (c *Editor) Unsubscribe(key uint64) {
	c.lock(func() {
		c.e.Unsubscribe(key)
	})
}
