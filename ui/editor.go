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
	if selector == nil {
		return nil
	}
	return selector
}

func getStatusStyle(mode string) tcell.Style {
	switch mode {
	case multimode_editor.ModeNormal:
		return tcell.StyleDefault.Background(tcell.ColorLightGray).Foreground(tcell.ColorBlack)
	case multimode_editor.ModeInsert:
		return tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack)
	case multimode_editor.ModeSelect:
		return tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)
	case multimode_editor.ModeCommand:
		return tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorBlack)
	default:
		return tcell.StyleDefault.Background(tcell.ColorLightGray).Foreground(tcell.ColorBlack)
	}
}

func getTextStyle(textRow int, selector *multimode_editor.Selector) tcell.Style {
	if selector != nil {
		beg, end := selector.Interval()
		if beg <= textRow && textRow <= end {
			return tcell.StyleDefault.Background(tcell.ColorLightGray).Foreground(tcell.ColorBlack)
		}
	}
	return tcell.StyleDefault
}

func draw(s tcell.Screen, view editor.View) {
	s.Clear()
	screenWidth, screenHeight := s.Size()
	selector := getSelector(view.Status.Other)

	// Draw cursor from (0, 0)
	s.ShowCursor(
		view.Cursor.Col-view.Window.TlCol,
		view.Cursor.Row-view.Window.TlRow,
	)

	// Draw content from (0, 0) -> (screenWidth-1, screenHeight-2)
	contentDrawContext := makeDrawContext(s, 0, 0, screenWidth, screenHeight-1)
	contentDrawContext(func(width int, height int, draw drawFunc) {
		t := view.Text
		for relRow := 0; relRow < height; relRow++ {
			row := view.Window.TlRow + relRow
			style := getTextStyle(row, selector)
			var line []rune = nil
			if row < t.Len() {
				line = t.Get(row)
			}

			for relCol := 0; relCol < width; relCol++ {
				col := view.Window.TlCol + relCol
				ch := ' '
				if col < len(line) {
					ch = line[col]
				}
				if row >= t.Len() && relCol == 0 {
					// special case
					ch = '~'
				}
				// s.SetContent(relCol, relRow, ch, nil, style)
				draw(relCol, relRow, ch, nil, style)
			}
		}
	})

	// Draw the status bar at the bottom (screenHeight-1)
	statusDrawContext := makeDrawContext(s, 0, screenHeight-1, screenWidth, 1)
	statusDrawContext(func(width int, height int, draw drawFunc) {
		mode, command := getModeAndCommand(view.Status.Other)
		style := getStatusStyle(mode)

		for col := 0; col < width; col++ {
			draw(col, 0, ' ', nil, style)
		}
		sep := []rune(" > ")
		var fromLeft []rune
		fromLeft = append(fromLeft, []rune(fmt.Sprintf(" %s (%d, %d)", mode, view.Cursor.Row+1, view.Cursor.Col+1))...)
		if len(command) > 0 {
			fromLeft = append(fromLeft, sep...)
			fromLeft = append(fromLeft, []rune(command)...)
		}
		if len(view.Status.Message) > 0 {
			fromLeft = append(fromLeft, sep...)
			fromLeft = append(fromLeft, []rune(view.Status.Message)...)
		}
		for col, ch := range fromLeft {
			draw(col, 0, ch, nil, style)
		}
		var fromRight []rune = nil
		if len(view.Status.Background) > 0 {
			fromRight = append(fromRight, sep...)
			fromRight = append(fromRight, []rune(view.Status.Background)...)
			fromRight = append(fromRight, ' ')
		}
		for i, ch := range fromRight {
			col := i + width - len(fromRight)
			draw(col, 0, ch, nil, style)
		}
	})

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
	insertEditor, loadCtx, finalizer, err := makeInsertEditor(ctx, inputFilename, logFilename, width, height-1)
	if err != nil {
		cancel()
		return err
	}
	defer finalizer.Close()

	if multiMode {
		e = multimode_editor.New(stop, insertEditor, inputFilename)
	} else {
		e = insertEditor
	}

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
				err := finalizer.Flush()
				if err != nil {
					writeMessage(e, fmt.Sprintf("flush error: %v", err))
				} else {
					writeMessage(e, "log_writer flushed")
				}
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
				err := finalizer.Flush()
				if err != nil {
					writeMessage(e, fmt.Sprintf("flush error: %v", err))
				} else {
					writeMessage(e, "log_writer flushed")
				}
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
		e.Action("mouse_click_left", editor.Cursor{
			Row: row,
			Col: col,
		})
	case button&tcell.Button2 != 0:
		e.Action("mouse_click_right", editor.Cursor{
			Row: row,
			Col: col,
		})
	case button&tcell.Button3 != 0:
		e.Action("mouse_click_right", editor.Cursor{
			Row: row,
			Col: col,
		})
	case button&tcell.Button4 != 0:
		e.Action("mouse_scroll_left")

	case button&tcell.Button5 != 0:
		e.Action("mouse_scroll_right")
	case button&tcell.WheelUp != 0:
		e.Action("mouse_scroll_up")

	case button&tcell.WheelDown != 0:
		e.Action("mouse_scroll_down")
	case button == tcell.ButtonNone:
		//e.Action("mouse_none", editor.Position{
		//		Row: row,
		//		Col: col,
		//	})
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
		e.Action("key_escape")
	case tcell.KeyTab:
		e.Action("key_tabular")

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
