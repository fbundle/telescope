package command_editor

import (
	"context"
	"fmt"
	"golang.org/x/exp/mmap"
	"strconv"
	"strings"
	"sync"
	"telescope/config"
	"telescope/editor"
	"telescope/exit"
	"telescope/log"
	"telescope/text"
	"time"
)

type Mode = string

const (
	ModeInsert  Mode = "INSERT"
	ModeCommand Mode = "COMMAND"
	ModeVisual  Mode = "VISUAL"
)

type commandEditor struct {
	mu               sync.Mutex
	mode             Mode
	e                editor.Editor
	command          []rune
	latestEditorView *editor.View
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
	writeMessage := func(msg string) {
		c.e.Message(msg)
		c.command = nil
		c.renderWithoutLock()
	}

	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeInsert:
			c.e.Enter()
		case ModeCommand:
			// apply command
			cmd := string(c.command)
			switch {
			case cmd == "i" || cmd == "insert":
				c.mode = ModeInsert
				c.command = nil
				c.renderWithoutLock()
			case strings.HasPrefix(cmd, "s ") || strings.HasPrefix(cmd, "search "):
				cmd = strings.TrimPrefix(cmd, "s ")
				cmd = strings.TrimPrefix(cmd, "search ")
				// TODO search
				exit.Write("search not implemented")
			case strings.HasPrefix(cmd, "g ") || strings.HasPrefix(cmd, "goto "):
				cmd = strings.TrimPrefix(cmd, "g ")
				cmd = strings.TrimPrefix(cmd, "goto ")
				lineNum, err := strconv.Atoi(cmd)
				if err != nil {
					writeMessage("invalid line number " + cmd)
					return
				}
				c.e.Goto(lineNum-1, 0)
				c.command = nil
				c.renderWithoutLock()
			default:
				c.e.Message("unknown command: " + string(c.command))
				c.command = nil
				c.renderWithoutLock()
				time.Sleep(config.Load().BLINKING_TIME)
				c.e.Message("")
				c.renderWithoutLock()
			}
		default:
			exit.Write("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Backspace() {
	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeInsert:
			c.e.Backspace()
		case ModeCommand:
			if len(c.command) > 0 {
				c.command = c.command[:len(c.command)-1]
			}
		default:
			exit.Write("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Delete() {
	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeInsert:
			c.e.Delete()
		case ModeCommand:
			// do nothing
		default:
			exit.Write("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Tabular() {
	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeInsert:
			c.e.Tabular()
		case ModeCommand:
		// do nothing
		default:
			exit.Write("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Goto(row int, col int) {
	c.lockUpdateRender(func() {
		c.e.Goto(row, col)
	})
}

func (c *commandEditor) MoveLeft() {
	c.lockUpdateRender(func() {
		c.e.MoveLeft()
	})
}

func (c *commandEditor) MoveRight() {
	c.lockUpdateRender(func() {
		c.e.MoveRight()
	})
}

func (c *commandEditor) MoveUp() {
	c.lockUpdateRender(func() {
		c.e.MoveUp()
	})
}

func (c *commandEditor) MoveDown() {
	c.lockUpdateRender(func() {
		c.e.MoveDown()
	})
}

func (c *commandEditor) MoveHome() {
	c.lockUpdateRender(func() {
		c.e.MoveHome()
	})
}

func (c *commandEditor) MoveEnd() {
	c.lockUpdateRender(func() {
		c.e.MoveEnd()
	})
}

func (c *commandEditor) MovePageUp() {
	c.lockUpdateRender(func() {
		c.e.MovePageUp()
	})
}

func (c *commandEditor) MovePageDown() {
	c.lockUpdateRender(func() {
		c.e.MovePageDown()
	})
}

func (c *commandEditor) Undo() {
	c.lockUpdateRender(func() {
		c.e.Undo()
	})
}

func (c *commandEditor) Redo() {
	c.lockUpdateRender(func() {
		c.e.Redo()
	})
}

func (c *commandEditor) Apply(entry log.Entry) {
	c.lockUpdateRender(func() {
		c.e.Apply(entry)
	})
}

func (c *commandEditor) Message(s string) {
	c.lockUpdateRender(func() {
		c.command = nil
		c.e.Message(s)
	})
}

func (c *commandEditor) Escape() {
	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeInsert:
			c.mode = ModeCommand
			c.command = nil
			c.renderWithoutLock()
		case ModeCommand:
		// do nothing
		default:
			exit.Write("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Text() text.Text {
	return c.e.Text()
}

func (c *commandEditor) renderWithoutLock() {
	if c.latestEditorView == nil {
		return
	}
	view := fromEditorView(*c.latestEditorView)
	view.Mode = c.mode
	if len(c.command) > 0 {
		view.Message = fmt.Sprintf("%s > %s", string(c.command), view.Message)
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

	f()
}

func NewCommandEditor(ctx context.Context, e editor.Editor) Editor {
	c := &commandEditor{
		mu:               sync.Mutex{},
		mode:             ModeVisual,
		e:                e,
		command:          nil,
		latestEditorView: nil,
		renderCh:         make(chan View, 1024),
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case view := <-c.e.Update():
				c.lockUpdateRender(func() {
					c.latestEditorView = &view
				})
			}
		}
	}()
	return c
}
