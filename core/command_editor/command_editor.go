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
	ModeNormal  Mode = "NORMAL"
	ModeCommand Mode = "COMMAND"
	ModeInsert  Mode = "INSERT"
	ModeSelect  Mode = "SELECT"
)

type Selector struct {
	Beg int
	End int
}

type clipboard [][]rune

type state struct {
	mode      Mode
	command   string
	selector  *Selector
	clipboard clipboard
}

type commandEditor struct {
	cancel func()
	mu     sync.Mutex
	e      editor.Editor
	state  state
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
		switch c.state.mode {
		case ModeNormal:
			switch ch {
			case 'i':
				c.enterInsertModeWithoutLock()
				c.writeWithoutLock("enter insert mode")
			case ':':
				c.enterCommandModeWithoutLock(":")
				c.writeWithoutLock("enter command mode")
			case 'V': // start selecting
				row := c.e.Render().Cursor.Row
				c.enterSelectModeWithoutLock(row)
				c.writeWithoutLock("enter select mode")
			case 'p': // paste
				if c.state.clipboard == nil {
					c.writeWithoutLock("clipboard is empty")
					return
				}
				row := c.e.Render().Cursor.Row
				c.e.InsertLine(row, c.state.clipboard)
				c.writeWithoutLock("pasted")
			default:
			}
		case ModeInsert:
			c.e.Type(ch)
		case ModeCommand:
			c.state.command += string(ch)
			c.writeWithoutLock("")
		case ModeSelect:
			switch ch {
			case 'd': // cut
				beg, end := c.state.selector.Beg, c.state.selector.End
				if beg > end {
					beg, end = end, beg
				}
				t := c.e.Render().Text
				clip := make([][]rune, 0)
				for i := beg; i <= end; i++ {
					clip = append(clip, t.Get(i))
				}
				c.state.clipboard = clip
				// delete
				c.e.DeleteLine(beg, len(c.state.clipboard))
				c.e.Goto(beg, 0)

				c.enterNormalModeWithoutLock()
				c.writeWithoutLock("cut")

			case 'y': //copy
				beg, end := c.state.selector.Beg, c.state.selector.End
				if beg > end {
					beg, end = end, beg
				}
				t := c.e.Render().Text
				clip := make([][]rune, 0)
				for i := beg; i <= end; i++ {
					clip = append(clip, t.Get(i))
				}
				c.state.clipboard = clip
				c.enterNormalModeWithoutLock()
				c.writeWithoutLock("copied")
			}
		default:
			side_channel.Panic("unknown mode: ", c.state)
		}
	})
}

func (c *commandEditor) applyCommandWithoutLock() {
	cmd := c.state.command

	switch {
	case cmd == ":i" || cmd == ":insert":
		c.enterInsertModeWithoutLock()
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
			c.enterNormalModeWithoutLock()
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
		c.enterNormalModeWithoutLock()
		c.writeWithoutLock("End of file")
		return
	case strings.HasPrefix(cmd, ":g ") || strings.HasPrefix(cmd, ":goto "):
		cmd = strings.TrimPrefix(cmd, ":g ")
		cmd = strings.TrimPrefix(cmd, ":goto ")
		lineNum, err := strconv.Atoi(cmd)
		if err != nil {
			c.enterNormalModeWithoutLock()
			c.writeWithoutLock("invalid line number " + cmd)
			return
		}
		c.e.Goto(lineNum-1, 0)
		c.enterNormalModeWithoutLock()
		c.writeWithoutLock("goto line " + cmd)
		return
	case strings.HasPrefix(cmd, ":w ") || strings.HasPrefix(cmd, ":writeWithoutLock "):
		cmd = strings.TrimPrefix(cmd, ":w ")
		cmd = strings.TrimPrefix(cmd, ":writeWithoutLock ")

		filename := cmd
		file, err := os.Create(filename)
		if err != nil {
			c.enterNormalModeWithoutLock()
			c.writeWithoutLock("error open file " + err.Error())
			return
		}
		defer file.Close()
		writer := bufio.NewWriter(file)
		for _, line := range c.e.Render().Text.Iter {
			_, err = writer.WriteString(string(line) + "\n")
			if err != nil {
				c.enterNormalModeWithoutLock()
				c.writeWithoutLock("error writeWithoutLock file " + err.Error())
				return
			}
		}
		err = writer.Flush()
		if err != nil {
			c.enterNormalModeWithoutLock()
			c.writeWithoutLock("error flush file " + err.Error())
			return
		}

		c.enterNormalModeWithoutLock()
		c.writeWithoutLock("file written into " + filename)
		return
	default:
		c.enterNormalModeWithoutLock()
		c.writeWithoutLock("unknown command: " + cmd)
	}
}

func (c *commandEditor) Enter() {
	c.lock(func() {
		switch c.state.mode {
		case ModeNormal:
			// do nothing
		case ModeInsert:
			c.e.Enter()
		case ModeCommand:
			c.applyCommandWithoutLock()
		case ModeSelect:
			// do nothing
		default:
			side_channel.Panic("unknown mode: ", c.state)
		}
	})
}

func (c *commandEditor) Escape() {
	c.lock(func() {
		c.enterNormalModeWithoutLock()
		c.writeWithoutLock("enter normal mode")
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
func (c *commandEditor) InsertLine(offset int, lines [][]rune) {
	c.lock(func() {
		c.e.InsertLine(offset, lines)
	})
}

func (c *commandEditor) DeleteLine(offset int, count int) {
	c.lock(func() {
		c.e.DeleteLine(offset, count)
	})
}

func NewCommandEditor(cancel func(), e editor.Editor) editor.Editor {
	c := &commandEditor{
		cancel: cancel,
		mu:     sync.Mutex{},
		e:      e,
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
