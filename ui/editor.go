package ui

import (
	"context"
	"errors"
	"fmt"
	"telescope/config"
	"telescope/core/command_editor"
	"telescope/core/editor"
	"telescope/util/side_channel"
	"time"

	"github.com/gdamore/tcell/v2"
)

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
	// Draw editor content from (0, 0)
	for row, line := range view.Window.Data {
		for col, ch := range line {
			s.SetContent(col, row, ch, nil, textStyle)
		}
	}
	// Draw cursor from (0, 0)
	s.ShowCursor(view.Window.Cursor.Col, view.Window.Cursor.Row)

	// Draw the status bar at the bottom
	for col := 0; col < screenWidth; col++ {
		s.SetContent(col, screenHeight-1, ' ', nil, statusStyle)
	}
	sep := []rune(" > ")
	fromLeft := []rune(fmt.Sprintf(" %s (%d, %d)", view.Status.Header, view.Cursor.Col+1, view.Cursor.Row+1))
	if len(view.Status.Command) > 0 {
		fromLeft = append(fromLeft, sep...)
		fromLeft = append(fromLeft, []rune(view.Status.Command)...)
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

func RunEditor(inputFilename string, logFilename string, commandMode bool) error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err = s.Init(); err != nil {
		return err
	}
	defer s.Fini()

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

	e, loadCtx, flush, closer, err := makeEditor(ctx, inputFilename, logFilename, width, height)
	if err != nil {
		cancel()
		return err
	}
	if commandMode {
		e = command_editor.NewCommandEditor(stop, e)
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
			// quit from editor
			running = false
		case *tcell.EventKey:
			if event.Key() == tcell.KeyCtrlC {
				// Ctrl+C to stop
				running = false
			} else if event.Key() == tcell.KeyCtrlS {
				// Ctrl+S to flush
				_ = flush()
				writeMessage(e, "log flushed")
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
	return nil
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
