package app

import (
	"context"
	"fmt"
	"time"

	"telescope/config"
	"telescope/editor"

	"github.com/gdamore/tcell/v2"
)

const (
	appName = "telescope"
)

func draw(s tcell.Screen, view editor.View) {
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
	fromLeft := []rune(fmt.Sprintf("(%d, %d)", view.TextCursor.Col, view.TextCursor.Row))
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

func RunEditor(inputFilename string, logFilename string) error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err = s.Init(); err != nil {
		return err
	}
	defer s.Fini()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	width, height := s.Size()

	e, loadCtx, flush, close, err := makeEditor(ctx, inputFilename, logFilename, width, height)
	if err != nil {
		return err
	}
	defer close()
	_ = loadCtx // do nothing with load ctx

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
		case *tcell.EventKey:
			if event.Key() == tcell.KeyCtrlC {
				// Ctrl+C to stop
				cancel()
				e.Message("stopping")
				// part of the code uses mmap.Reader
				// even though it check for ctx.Done()
				// but if we close the file too soon, the reading is still on going
				// hence we will wait for a while before closing the file
				time.Sleep(time.Second)
				running = false
			} else if event.Key() == tcell.KeyCtrlS {
				// Ctrl+S to flush
				_ = flush()
				e.Message("log flushed")
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
		// e.Escape() // TODO - uncomment for CommandEditor
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
