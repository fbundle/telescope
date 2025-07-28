package app

import (
	"context"
	"fmt"
	"telescope/command_editor"
	"time"

	"telescope/config"

	"github.com/gdamore/tcell/v2"
)

func drawCommandEditor(s tcell.Screen, view command_editor.View) {
	statusStyle := tcell.StyleDefault.
		Background(tcell.ColorLightGray).
		Foreground(tcell.ColorBlack)
	textStyle := tcell.StyleDefault

	s.Clear()
	screenWidth, screenHeight := s.Size()
	// Draw editor content from (0, 0)
	for row, line := range view.WinData {
		for col, ch := range line {
			s.SetContent(col, row, ch, nil, textStyle)
		}
	}
	// Draw cursor from (0, 0)
	s.ShowCursor(view.WinCursor.Col, view.WinCursor.Row)

	// Draw the status bar at the bottom
	for col := 0; col < screenWidth; col++ {
		s.SetContent(col, screenHeight-1, ' ', nil, statusStyle)
	}
	sep := []rune(" > ")
	fromLeft := []rune{' '}
	fromLeft = append(fromLeft, []rune(fmt.Sprintf("%s (%d, %d)", view.Mode, view.TextCursor.Col+1, view.TextCursor.Row+1))...)
	fromLeft = append(fromLeft, sep...)
	fromLeft = append(fromLeft, []rune(view.Message)...)
	for col, ch := range fromLeft {
		if 0 <= col && col < screenWidth {
			s.SetContent(col, screenHeight-1, ch, nil, statusStyle)
		}
	}
	fromRight := append(sep, []rune(view.Background)...)
	for i, ch := range fromRight {
		col := i + screenWidth - len(fromRight)
		if 0 <= col && col < screenWidth {
			s.SetContent(col, screenHeight-1, ch, nil, statusStyle)
		}
	}

	s.Show()
}

func RunCommandEditor(inputFilename string, logFilename string) error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err = s.Init(); err != nil {
		return err
	}
	defer s.Fini()

	ctx, cancel := context.WithCancel(context.Background())
	width, height := s.Size()

	e1, loadCtx, flush, closer, err := makeEditor(ctx, inputFilename, logFilename, width, height)
	e := command_editor.NewCommandEditor(ctx, e1)
	if err != nil {
		cancel()
		return err
	}
	defer closer()

	// draw loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case view := <-e.Update():
				drawCommandEditor(s, view)
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
		case *tcell.EventKey:
			if event.Key() == tcell.KeyCtrlC {
				// Ctrl+C to stop
				running = false
			} else if event.Key() == tcell.KeyCtrlS {
				// Ctrl+S to flush
				_ = flush()
				e.Message("log flushed")
			} else {
				handleCommandEditorKey(e, event)
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
	e.Message("exiting... ")
	cancel()
	<-loadCtx.Done() // wait for load context then exit, exec deferred closer function
	return nil
}

func handleCommandEditorKey(e command_editor.Editor, ev *tcell.EventKey) {
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
		e.Message(fmt.Sprintf("unknown key %v", ev.Name()))
	}
}
