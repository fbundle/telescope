package command_editor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"telescope/editor"
	"telescope/side_channel"

	"telescope/log"
	"telescope/text"

	"golang.org/x/exp/mmap"
)

type Mode = string

const (
	ModeInsert  Mode = "INSERT"
	ModeCommand Mode = "COMMAND"
	ModeVisual  Mode = "VISUAL"
)

type commandEditor struct {
	mu                 sync.Mutex
	mode               Mode
	e                  editor.Editor
	command            []rune
	firstEditorViewCtx context.Context
	latestEditorView   *atomic.Value // editor.View
	renderCh           chan View
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

// NOTE - begins with VISUAL mode, type : to enter COMMAND mode
// press enter in COMMAND mode to apply command, if command is :i or :insert, enter insert mode
// otherwise, enter COMMAND mode
// press esc in any mode go back to VISUAL mode

func (c *commandEditor) Type(ch rune) {
	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeVisual:
			if ch == ':' {
				c.mode = ModeCommand
				c.command = []rune{':'}
				c.renderWithoutLock()
			}
		case ModeInsert:
			c.e.Type(ch)
		case ModeCommand:
			c.command = append(c.command, ch)
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func applyCommand(command []rune, c *commandEditor) ([]rune, Mode, string) {
	cmd := string(command)
	cmd = strings.TrimSpace(cmd)

	switch {
	case cmd == ":i" || cmd == ":insert":
		return nil, ModeInsert, ""

	case strings.HasPrefix(cmd, ":s ") || strings.HasPrefix(cmd, ":search "):
		cmd = strings.TrimPrefix(cmd, ":s ")
		cmd = strings.TrimPrefix(cmd, ":search ")

		text := c.e.Text()
		row := c.latestEditorView.Load().(editor.View).TextCursor.Row
		_, text2 := text.Split(row)

		for i, line := range text2.Iter { // TODO check error here
			if strings.Contains(string(line), cmd) {
				c.e.Goto(i, 0)
				return command, ModeCommand, ""
			}
		}

		return nil, ModeVisual, "substring not found"

	case strings.HasPrefix(cmd, ":g ") || strings.HasPrefix(cmd, ":goto "):
		cmd = strings.TrimPrefix(cmd, ":g ")
		cmd = strings.TrimPrefix(cmd, ":goto ")
		lineNum, err := strconv.Atoi(cmd)
		if err != nil {
			return nil, ModeVisual, "invalid line number " + cmd
		}
		c.e.Goto(lineNum-1, 0)
		return nil, ModeVisual, ""

	case strings.HasPrefix(cmd, ":w ") || strings.HasPrefix(cmd, ":write "):
		cmd = strings.TrimPrefix(cmd, ":w ")
		cmd = strings.TrimPrefix(cmd, ":write ")

		filename := cmd
		text := c.e.Text()
		file, err := os.Create(filename)
		if err != nil {
			return nil, ModeVisual, "error open file " + err.Error()
		}
		defer file.Close()
		writer := bufio.NewWriter(file)
		for _, line := range text.Iter {
			writer.WriteString(string(line) + "\n")
		}
		err = writer.Flush()
		if err != nil {
			return nil, ModeVisual, "error flush file " + err.Error()
		}

		return nil, ModeVisual, "file written into " + filename
	default:
		return nil, ModeVisual, "unknown command: " + cmd
	}

}

func (c *commandEditor) Enter() {

	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeVisual:
			// do nothing
		case ModeInsert:
			c.e.Enter()
		case ModeCommand:
			command, mode, message := applyCommand(c.command, c)
			c.command = command
			c.mode = mode
			c.e.WriteMessage(message)
			c.renderWithoutLock()
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Escape() {
	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeVisual:
			// do nothing
		case ModeInsert:
			c.mode = ModeVisual
			c.renderWithoutLock()
		case ModeCommand:
			c.mode = ModeVisual
			c.command = nil
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Backspace() {
	c.lockUpdateRender(func() {
		switch c.mode {
		case ModeVisual:
			// do nothing
		case ModeInsert:
			c.e.Backspace()
		case ModeCommand:
			if len(c.command) > 0 {
				c.command = c.command[:len(c.command)-1]
			}
		default:
			side_channel.Panic("unknown mode: ", c.mode)
		}
	})
}

func (c *commandEditor) Delete() {
	c.lockUpdateRender(func() {
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
	c.lockUpdateRender(func() {
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
		c.e.WriteMessage(s)
	})
}

func (c *commandEditor) Text() text.Text {
	return c.e.Text()
}

func (c *commandEditor) getView() View {
	<-c.firstEditorViewCtx.Done()
	view := c.latestEditorView.Load().(editor.View)

	return View{
		Mode:       "",
		WinData:    view.WinData,
		WinCursor:  Cursor{Row: view.WinCursor.Row, Col: view.WinCursor.Col},
		TextCursor: Cursor{Row: view.TextCursor.Row, Col: view.TextCursor.Col},
		Message:    view.Message,
		Background: view.Background,
	}
}

func (c *commandEditor) renderWithoutLock() {
	view := c.getView()
	view.Mode = c.mode
	if len(c.command) > 0 {
		view.Message = fmt.Sprintf("%s > %s", string(c.command), view.Message)
	}
	c.renderCh <- view
}

func (c *commandEditor) lockUpdateRender(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.renderWithoutLock()

	f()
}

func NewCommandEditor(ctx context.Context, e editor.Editor) Editor {
	firstEditorViewCtx, cancel := context.WithCancel(ctx)
	c := &commandEditor{
		mu:                 sync.Mutex{},
		mode:               ModeVisual,
		e:                  e,
		command:            nil,
		firstEditorViewCtx: firstEditorViewCtx,
		latestEditorView:   &atomic.Value{},
		renderCh:           make(chan View, 1),
	}

	c.latestEditorView.Store(editor.View{})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case view := <-c.e.Update():
				c.latestEditorView.Store(view)
				cancel()
				// c.lockUpdateRender(func() {}) // TODO - enable it seems to make it hang
			}
		}
	}()
	return c
}
