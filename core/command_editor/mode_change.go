package command_editor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"telescope/config"
	seq "telescope/util/persistent/sequence"
	"telescope/util/side_channel"
	"time"
)

// all functions resulting in mode change

func (c *commandEditor) Escape() {
	c.lock(func() {
		c.enterNormalModeWithoutLock()
		c.writeWithoutLock("enter normal mode")
	})
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
				c.e.InsertLine(c.state.clipboard)
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
				l := c.e.Render().Text.Lines
				c.state.clipboard = seq.Slice(l, beg, end+1)
				// delete
				c.e.Goto(beg, 0)
				c.e.DeleteLine(c.state.clipboard.Len())

				c.enterNormalModeWithoutLock()
				c.writeWithoutLock("cut")

			case 'y': //copy
				beg, end := c.state.selector.Beg, c.state.selector.End
				if beg > end {
					beg, end = end, beg
				}
				l := c.e.Render().Text.Lines
				c.state.clipboard = seq.Slice(l, beg, end+1)
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
		return

	case strings.HasPrefix(cmd, ":s ") || strings.HasPrefix(cmd, ":search ") || strings.HasPrefix(cmd, ":regex "):
		regex := strings.HasPrefix(cmd, ":regex ")
		cmd = strings.TrimPrefix(cmd, ":s ")
		cmd = strings.TrimPrefix(cmd, ":search ")
		cmd = strings.TrimPrefix(cmd, ":regex ")

		var match func(line string) bool
		if regex {
			re, err := regexp.Compile(cmd)
			if err != nil {
				c.enterNormalModeWithoutLock()
				c.writeWithoutLock(fmt.Sprintf("regexp compile error %s", err.Error()))
				return
			}
			match = re.MatchString
		} else {
			match = func(line string) bool {
				return strings.Contains(line, cmd)
			}
		}

		view := c.e.Render()
		row := view.Cursor.Row

		t := view.Text
		lines := seq.Slice(t.Lines, row+1, t.Lines.Len())

		t0 := time.Now()
		for i, l := range lines.Iter {
			line := l.Repr(t.Reader)
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
		c.writeWithoutLock("end of file")
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
	case cmd == ":w" || cmd == ":write" || cmd == ":wq":
		filename := c.defaultOutputFile

		// write file
		err := safeWriteFile(filename, c.e.Render().Text.Iter)
		if err != nil {
			c.enterNormalModeWithoutLock()
			c.writeWithoutLock("error write file " + err.Error())
			return
		}
		c.enterNormalModeWithoutLock()
		c.writeWithoutLock("file written into " + filename)

		// delete log and exit
		c.cancel()
		return

	case strings.HasPrefix(cmd, ":w ") || strings.HasPrefix(cmd, ":write "):
		cmd = strings.TrimPrefix(cmd, ":w ")
		cmd = strings.TrimPrefix(cmd, ":write ")

		filename := strings.TrimSpace(cmd)

		// write file
		err := safeWriteFile(filename, c.e.Render().Text.Iter)
		if err != nil {
			c.enterNormalModeWithoutLock()
			c.writeWithoutLock("error write file " + err.Error())
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
