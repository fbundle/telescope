package command_editor

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"telescope/bytes"
	"telescope/editor"
	"telescope/side_channel"

	"telescope/log"
	"telescope/text"
)

type Mode = string

const (
	ModeInsert  Mode = "INSERT"
	ModeCommand Mode = "COMMAND"
	ModeVisual  Mode = "VISUAL"
)

type commandEditor struct {
	mu      sync.Mutex
	mode    Mode
	e       editor.Editor
	command string
}

func (c *commandEditor) Update() <-chan editor.View {
	return c.e.Update()
}

func (c *commandEditor) Load(ctx context.Context, reader bytes.Array) (loadCtx context.Context, err error) {
	c.lockUpdate(func() {
		loadCtx, err = c.e.Load(ctx, reader)
	})
	return loadCtx, err
}

func (c *commandEditor) Resize(height int, width int) {
	c.lockUpdate(func() {
		c.e.Resize(height, width)
	})
}

// NOTE - begins with VISUAL mode, type : to enter COMMAND mode
// press enter in COMMAND mode to apply command, if command is :i or :insert, enter insert mode
// otherwise, enter COMMAND mode
// press esc in any mode go back to VISUAL mode

func (c *commandEditor) Type(ch rune) {
	c.lockUpdate(func() {
		switch c.mode {
		case ModeVisual:
			switch ch {
			case 'i':
				c.mode = ModeInsert
				c.e.WriteHeaderAndMessage(c.mode, "enter insert mode")
			case ':':
				c.mode = ModeCommand
				c.command = string(ch)
				c.e.WriteHeaderAndMessage(c.mode, c.command)
			default:
			}
		case ModeInsert:
			c.e.Type(ch)
		case ModeCommand:
			c.command += string(ch)
			c.e.WriteHeaderAndMessage(c.mode, c.command)
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func applyCommand(command string, c *commandEditor) (nextCommand string, nextMode Mode, message string) {
	cmd := command
	cmd = strings.TrimSpace(cmd)

	switch {
	case cmd == ":i" || cmd == ":insert":
		return "", ModeInsert, ""

	case strings.HasPrefix(cmd, ":s ") || strings.HasPrefix(cmd, ":search "):
		cmd = strings.TrimPrefix(cmd, ":s ")
		cmd = strings.TrimPrefix(cmd, ":search ")

		side_channel.Write(cmd)

		_, text2 := c.e.Text().Split(c.e.Cursor().Row)

		for i, line := range text2.Iter {
			if strings.Contains(string(line), cmd) {
				c.e.Goto(i, 0)
				return command, ModeCommand, ""
			}
		}

		return "", ModeVisual, "substring not found"

	case strings.HasPrefix(cmd, ":g ") || strings.HasPrefix(cmd, ":goto "):
		cmd = strings.TrimPrefix(cmd, ":g ")
		cmd = strings.TrimPrefix(cmd, ":goto ")
		lineNum, err := strconv.Atoi(cmd)
		if err != nil {
			return "", ModeVisual, "invalid line number " + cmd
		}
		c.e.Goto(lineNum-1, 0)
		return "", ModeVisual, ""

	case strings.HasPrefix(cmd, ":w ") || strings.HasPrefix(cmd, ":write "):
		cmd = strings.TrimPrefix(cmd, ":w ")
		cmd = strings.TrimPrefix(cmd, ":write ")

		filename := cmd
		file, err := os.Create(filename)
		if err != nil {
			return "", ModeVisual, "error open file " + err.Error()
		}
		defer file.Close()
		writer := bufio.NewWriter(file)
		for _, line := range c.e.Text().Iter {
			_, err = writer.WriteString(string(line) + "\n")
			if err != nil {
				return "", ModeVisual, "error write file " + err.Error()
			}
		}
		err = writer.Flush()
		if err != nil {
			return "", ModeVisual, "error flush file " + err.Error()
		}

		return "", ModeVisual, "file written into " + filename
	default:
		return "", ModeVisual, "unknown command: " + cmd
	}

}

func (c *commandEditor) Enter() {
	c.lockUpdate(func() {
		switch c.mode {
		case ModeVisual:
			// do nothing
		case ModeInsert:
			c.e.Enter()
		case ModeCommand:
			nextCommand, nextMode, message := applyCommand(c.command, c)
			c.command = nextCommand
			c.mode = nextMode
			if len(c.command) > 0 {
				c.e.WriteHeaderAndMessage(c.mode, c.command)
			} else {
				c.e.WriteHeaderAndMessage(c.mode, message)
			}
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Escape() {
	c.lockUpdate(func() {
		switch c.mode {
		case ModeVisual:
			// do nothing
		case ModeInsert:
			c.mode = ModeVisual
			c.e.WriteHeaderAndMessage(c.mode, "enter visual mode")
		case ModeCommand:
			c.mode = ModeVisual
			c.command = ""
			c.e.WriteHeaderAndMessage(c.mode, "enter visual mode")
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Backspace() {
	c.lockUpdate(func() {
		switch c.mode {
		case ModeVisual:
			// do nothing
		case ModeInsert:
			c.e.Backspace()
		case ModeCommand:
			if len(c.command) > 0 {
				c.command = c.command[:len(c.command)-1]
			}
			c.e.WriteHeaderAndMessage(c.mode, c.command)
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Delete() {
	c.lockUpdate(func() {
		switch c.mode {
		case ModeVisual:
			// do nothing
		case ModeInsert:
			c.e.Delete()
		case ModeCommand:
			// do nothing
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Tabular() {
	c.lockUpdate(func() {
		switch c.mode {
		case ModeVisual:
			// do nothing
		case ModeInsert:
			c.e.Tabular()
		case ModeCommand:
			// do nothing
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Goto(row int, col int) {
	c.lockUpdate(func() {
		c.e.Goto(row, col)
	})
}

func (c *commandEditor) MoveLeft() {
	c.lockUpdate(func() {
		c.e.MoveLeft()
	})
}

func (c *commandEditor) MoveRight() {
	c.lockUpdate(func() {
		c.e.MoveRight()
	})
}

func (c *commandEditor) MoveUp() {
	c.lockUpdate(func() {
		c.e.MoveUp()
	})
}

func (c *commandEditor) MoveDown() {
	c.lockUpdate(func() {
		c.e.MoveDown()
	})
}

func (c *commandEditor) MoveHome() {
	c.lockUpdate(func() {
		c.e.MoveHome()
	})
}

func (c *commandEditor) MoveEnd() {
	c.lockUpdate(func() {
		c.e.MoveEnd()
	})
}

func (c *commandEditor) MovePageUp() {
	c.lockUpdate(func() {
		c.e.MovePageUp()
	})
}

func (c *commandEditor) MovePageDown() {
	c.lockUpdate(func() {
		c.e.MovePageDown()
	})
}

func (c *commandEditor) Undo() {
	c.lockUpdate(func() {
		c.e.Undo()
	})
}

func (c *commandEditor) Redo() {
	c.lockUpdate(func() {
		c.e.Redo()
	})
}

func (c *commandEditor) Apply(entry log.Entry) {
	c.lockUpdate(func() {
		c.e.Apply(entry)
	})
}

func (c *commandEditor) WriteHeaderAndMessage(header string, message string) {
	c.lockUpdate(func() {
		c.e.WriteHeaderAndMessage(header, message)
	})
}

func (c *commandEditor) WriteMessage(message string) {
	c.lockUpdate(func() {
		c.e.WriteMessage(message)
	})
}

func (c *commandEditor) Text() text.Text {
	return c.e.Text()
}
func (c *commandEditor) Cursor() editor.Cursor {
	return c.e.Cursor()
}

func (c *commandEditor) lockUpdate(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	f()
}

func NewCommandEditor(e editor.Editor) editor.Editor {
	c := &commandEditor{
		mu:      sync.Mutex{},
		mode:    ModeVisual,
		e:       e,
		command: "",
	}
	c.e.WriteHeaderAndMessage(c.mode, "")
	return c
}
