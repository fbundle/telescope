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
	"telescope/config"
	"telescope/core/editor"
	"telescope/core/log"
	"telescope/util/bytes"
	"telescope/util/side_channel"
	"time"
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

func (c *commandEditor) Type(ch rune) {
	c.lock(func() {
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
	c.lock(func() {
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
	c.lock(func() {
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
	c.lock(func() {
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
	c.lock(func() {
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
	c.lock(func() {
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
	})
}

func (c *commandEditor) MoveDown() {
	c.lock(func() {
		c.e.MoveDown()
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
	})
}

func (c *commandEditor) MovePageDown() {
	c.lock(func() {
		c.e.MovePageDown()
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

func (c *commandEditor) lock(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	f()
}

func (c *commandEditor) writeWithoutLock(message string) {
	c.e.Status(func(status editor.Status) editor.Status {
		status.Header = c.mode
		status.Command = c.command
		status.Message = message
		return status
	})
}
