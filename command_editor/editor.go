package command_editor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"telescope/bytes"
	"telescope/config"
	"telescope/editor"
	"telescope/side_channel"
	"time"

	"telescope/log"
)

type Mode = string

const (
	ModeInsert  Mode = "INSERT"
	ModeCommand Mode = "COMMAND"
	ModeVisual  Mode = "VISUAL"
)

type commandEditor struct {
	cancel  func()
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
				c.mode, c.command = ModeInsert, ""
				c.writeWithoutLock("enter insert mode")
			case ':':
				c.mode, c.command = ModeCommand, string(ch)
				c.writeWithoutLock("enter command mode")
			default:
			}
		case ModeInsert:
			c.e.Type(ch)
		case ModeCommand:
			c.command += string(ch)
			c.writeWithoutLock("")
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) applyCommandWithoutLock() {
	cmd := c.command

	switch {
	case cmd == ":i" || cmd == ":insert":
		c.mode, c.command = ModeInsert, ""
		c.writeWithoutLock("enter insert mode")
		return

	case cmd == ":q" || cmd == ":quit":
		c.cancel()
		c.writeWithoutLock("exitting ...")
		return

	case strings.HasPrefix(cmd, ":s ") || strings.HasPrefix(cmd, ":search ") || strings.HasPrefix(cmd, ":regex "):
		regex := strings.HasPrefix(cmd, ":regex ")
		cmd = strings.TrimPrefix(cmd, ":s ")
		cmd = strings.TrimPrefix(cmd, ":search ")
		cmd = strings.TrimPrefix(cmd, ":regex ")

		re, err := regexp.Compile(cmd)
		if err != nil {
			c.mode, c.command = ModeVisual, ""
			c.writeWithoutLock(fmt.Sprintf("regexp compile error %s", err.Error()))
			return
		}
		var match func(line string) bool
		if regex {
			match = re.MatchString
		} else {
			match = func(line string) bool {
				return strings.Contains(line, cmd)
			}
		}

		view := c.e.Render()
		row := view.Cursor.Row

		_, text2 := view.Text.Split(row + 1)

		t0 := time.Now()
		for i, line := range text2.Iter {
			if match(string(line)) {
				c.e.Goto(row+1+i, 0)
				c.writeWithoutLock("found substring " + cmd)
				return
			}
			t1 := time.Now()
			if t1.Sub(t0) > config.Load().MAX_SEACH_TIME {
				c.writeWithoutLock(fmt.Sprintf("search timeout after %d seconds and %d entries", config.Load().MAX_SEACH_TIME/time.Second, i+1))
				return
			}
		}

		c.mode, c.command = ModeVisual, ""
		c.writeWithoutLock("end of file")
		return
	case strings.HasPrefix(cmd, ":g ") || strings.HasPrefix(cmd, ":goto "):
		cmd = strings.TrimPrefix(cmd, ":g ")
		cmd = strings.TrimPrefix(cmd, ":goto ")
		lineNum, err := strconv.Atoi(cmd)
		if err != nil {
			c.mode, c.command = ModeVisual, ""
			c.writeWithoutLock("invalid line number " + cmd)
			return
		}
		c.e.Goto(lineNum-1, 0)
		c.mode, c.command = ModeVisual, ""
		c.writeWithoutLock("goto line " + cmd)
		return
	case strings.HasPrefix(cmd, ":w ") || strings.HasPrefix(cmd, ":writeWithoutLock "):
		cmd = strings.TrimPrefix(cmd, ":w ")
		cmd = strings.TrimPrefix(cmd, ":writeWithoutLock ")

		filename := cmd
		file, err := os.Create(filename)
		if err != nil {
			c.mode, c.command = ModeVisual, ""
			c.writeWithoutLock("error open file " + err.Error())
			return
		}
		defer file.Close()
		writer := bufio.NewWriter(file)
		for _, line := range c.e.Render().Text.Iter {
			_, err = writer.WriteString(string(line) + "\n")
			if err != nil {
				c.mode, c.command = ModeVisual, ""
				c.writeWithoutLock("error writeWithoutLock file " + err.Error())
				return
			}
		}
		err = writer.Flush()
		if err != nil {
			c.mode, c.command = ModeVisual, ""
			c.writeWithoutLock("error flush file " + err.Error())
			return
		}

		c.mode, c.command = ModeVisual, ""
		c.writeWithoutLock("file written into " + filename)
		return
	default:
		c.mode, c.command = ModeVisual, ""
		c.writeWithoutLock("unknown command: " + cmd)
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
			c.applyCommandWithoutLock()
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
			c.mode, c.command = ModeVisual, ""
			c.writeWithoutLock("enter visual mode")
		case ModeCommand:
			c.mode, c.command = ModeVisual, ""
			c.writeWithoutLock("enter visual mode")
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
			c.writeWithoutLock("")
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

func (c *commandEditor) WriteHeaderCommandMessage(header string, command string, message string) {
	c.lockUpdate(func() {
		writeHeaderCommandMessage(c.e, header, command, message)
	})
}

func (c *commandEditor) Render() editor.View {
	return c.e.Render()
}

func (c *commandEditor) lockUpdate(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	f()
}

func (c *commandEditor) writeWithoutLock(message string) {
	writeHeaderCommandMessage(c.e, c.mode, c.command, message)
}

func (c *commandEditor) Status(update func(status editor.Status) editor.Status) {
	c.e.Status(update)
}

func NewCommandEditor(cancel func(), e editor.Editor) editor.Editor {
	c := &commandEditor{
		cancel:  cancel,
		mu:      sync.Mutex{},
		mode:    ModeVisual,
		e:       e,
		command: "",
	}
	c.writeWithoutLock("")
	return c
}

func writeHeaderCommandMessage(e editor.Editor, header string, command string, message string) {
	e.Status(func(status editor.Status) editor.Status {
		status.Header = header
		status.Command = command
		status.Message = message
		return status
	})
}
