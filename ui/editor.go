package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"telescope/config"
	"telescope/core/editor"
	"telescope/core/multimode_editor"
	"telescope/util/side_channel"
	"time"

	"github.com/gdamore/tcell/v2"
)

func getModeAndCommand(m map[string]any) (string, string) {
	if m == nil {
		return "", ""
	}
	mode := ""
	s, ok := m["mode"]
	if ok {
		mode = fmt.Sprintf("%v", s)
	}
	command := ""
	s, ok = m["command"]
	if ok {
		command = fmt.Sprintf("%v", s)
	}
	return mode, command
}

func getSelector(m map[string]any) *multimode_editor.Selector {
	if m == nil {
		return nil
	}
	s, ok := m["selector"]
	if !ok {
		return nil
	}
	selector, ok := s.(*multimode_editor.Selector)
	if !ok {
		return nil
	}
	return selector
}

func draw(s tcell.Screen, view editor.View) {
	statusStyle := tcell.StyleDefault.
		Background(tcell.ColorLightGray).
		Foreground(tcell.ColorBlack)
	textStyle := tcell.StyleDefault
	highlightStyle := tcell.StyleDefault.
		Background(tcell.ColorLightGray).
		Foreground(tcell.ColorBlack)

	_ = highlightStyle

	s.Clear()
	screenWidth, screenHeight := s.Size()
	selector := getSelector(view.Status.Other)
	// Draw insert_editor content from (0, 0)
	for row, line := range view.Window.Data {
		style := textStyle
		if selector != nil {
			beg, end := selector.Beg, selector.End
			if beg > end {
				beg, end = end, beg
			}
			textRow := view.Window.TopLeft.Row + row
			if beg <= textRow && textRow <= end {
				style = highlightStyle
			}
		}
		for col, ch := range line {
			s.SetContent(col, row, ch, nil, style)
		}
	}
	// Draw cursor from (0, 0)
	relRow := view.Cursor.Row - view.Window.TopLeft.Row
	relCol := view.Cursor.Col - view.Window.TopLeft.Col
	s.ShowCursor(relCol, relRow)

	// Draw the status bar at the bottom
	for col := 0; col < screenWidth; col++ {
		s.SetContent(col, screenHeight-1, ' ', nil, statusStyle)
	}

	header, command := getModeAndCommand(view.Status.Other)
	sep := []rune(" > ")
	fromLeft := []rune(fmt.Sprintf(" %s (%d, %d)", header, view.Cursor.Col+1, view.Cursor.Row+1))
	if len(command) > 0 {
		fromLeft = append(fromLeft, sep...)
		fromLeft = append(fromLeft, []rune(command)...)
	}
	if len(view.Status.Message) > 0 {
		fromLeft = append(fromLeft, sep...)
		fromLeft = append(fromLeft, []rune(view.Status.Message)...)
	}
	for col, ch := range fromLeft {
		if 0 <= col && col < screenWidth {
			s.SetContent(col, screenHeight-1, ch, nil, statusStyle)
		}
	}
	fromRight := []rune{}
	if len(view.Status.Background) > 0 {
		fromRight = append(fromRight, sep...)
		fromRight = append(fromRight, []rune(view.Status.Background)...)
	}
	for i, ch := range fromRight {
		col := i + screenWidth - len(fromRight)
		if 0 <= col && col < screenWidth {
			s.SetContent(col, screenHeight-1, ch, nil, statusStyle)
		}
	}

	s.Show()
}

type quitEvent struct {
	when time.Time
}

func (e *quitEvent) When() time.Time {
	return e.when
}

func RunEditor(inputFilename string, logFilename string, multiMode bool) error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err = s.Init(); err != nil {
		return err
	}
	defer s.Fini()

	s.EnableMouse()

	ctx, cancel := context.WithCancel(context.Background())
	stop := func() {
		cancel()
		var err error
		for i := 0; i < 5; i++ {
			err = s.PostEvent(&quitEvent{when: time.Now()})
			if err == nil {
				break
			}
			if !errors.Is(err, tcell.ErrEventQFull) {
				side_channel.Panic(err)
			}
			time.Sleep(100 * time.Millisecond) // retry stopping
		}
	}

	width, height := s.Size()

	var e editor.Editor
	// make editor
	insertEditor, loadCtx, flush, closer, err := makeInsertEditor(ctx, inputFilename, logFilename, width, height)
	if err != nil {
		cancel()
		return err
	}
	if multiMode {
		multiModeEditor := multimode_editor.New(stop, insertEditor, inputFilename)
		e = multiModeEditor
	} else {
		e = insertEditor
	}
	defer closer()

	// draw loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case view := <-e.Update():
				draw(s, view)
			}
		}
	}()
	// manual flush loop in the event of crash
	go func() {
		ticker := time.NewTicker(config.Load().LOG_AUTOFLUSH_INTERVAL)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				flush()
			}
		}
	}()

	// event loop
	running := true
	for running {
		event := s.PollEvent()
		switch event := event.(type) {
		case *quitEvent:
			// quit from insert_editor
			running = false
		case *tcell.EventMouse:
			handleEditorMouse(e, event)
		case *tcell.EventKey:
			if event.Key() == tcell.KeyCtrlC {
				// Ctrl+C to stop
				running = false
			} else if event.Key() == tcell.KeyCtrlS {
				// Ctrl+S to flush
				_ = flush()
				writeMessage(e, "log_writer flushed")
			} else {
				handleEditorKey(e, event)
			}

		case *tcell.EventResize:
			s.Sync()
			width, height = s.Size()
			e.Resize(height-1, width)
		default:
			// nothing
		}
	}

	// we have to cancel here first, then wait for a while before exiting
	// since exiting will close all the files; waiting time is necessary for all background tasks to stop reading files
	writeMessage(e, "exiting... ")
	cancel()
	<-loadCtx.Done() // wait for load context then exit, exec deferred closer function

	s.Clear() // last thing to do - clear screen, delete log_writer file
	_ = os.Remove(logFilename)
	return nil
}

func handleEditorMouse(e editor.Editor, ev *tcell.EventMouse) {
	col, row := ev.Position()
	button := ev.Buttons()
	switch {
	case button&tcell.Button1 != 0:
		// msg = fmt.Sprintf("Left click at (%d, %d)", col, row)
		e.Action(map[string]any{
			"mouse_click": editor.Position{
				Row: row,
				Col: col,
			},
		})
	case button&tcell.Button2 != 0:
		//msg = fmt.Sprintf("Middle click at (%d, %d)", col, row)
	case button&tcell.Button3 != 0:
		//msg = fmt.Sprintf("Right click at (%d, %d)", col, row)
	case button&tcell.WheelUp != 0:
		//msg = "Mouse wheel up"
		e.Action(map[string]any{
			"mouse_scroll_up": nil,
		})

	case button&tcell.WheelDown != 0:
		//msg = "Mouse wheel down"
		e.Action(map[string]any{
			"mouse_scroll_down": nil,
		})
	case button == tcell.ButtonNone:
		// msg = fmt.Sprintf("Mouse move at (%d, %d)", col, row)
	}
}

func handleEditorKey(e editor.Editor, ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyRune:
		e.Type(ev.Rune())
	case tcell.KeyEnter:
		e.Enter()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.Backspace()
	case tcell.KeyDelete:
		e.Delete()

	case tcell.KeyEsc:
		e.Escape()
	case tcell.KeyTab:
		e.Tabular()

	case tcell.KeyRight:
		e.MoveRight()
	case tcell.KeyLeft:
		e.MoveLeft()
	case tcell.KeyUp:
		e.MoveUp()
	case tcell.KeyDown:
		e.MoveDown()
	case tcell.KeyHome:
		e.MoveHome()
	case tcell.KeyEnd:
		e.MoveEnd()
	case tcell.KeyPgUp:
		e.MovePageUp()
	case tcell.KeyPgDn:
		e.MovePageDown()
	case tcell.KeyCtrlU:
		e.Undo()
	case tcell.KeyCtrlR:
		e.Redo()
	default:
		writeMessage(e, fmt.Sprintf("unknown key %v", ev.Name()))
	}
}
