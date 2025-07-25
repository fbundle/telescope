package app

import (
	"context"
	"fmt"
	"time"

	"os"
	"telescope/editor"
	"telescope/feature"

	"github.com/gdamore/tcell/v2"
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

	// Draw status bar at the bottom
	head := []rune(fmt.Sprintf("%s (%d, %d)", view.WinName, view.TextCursor.Col, view.TextCursor.Row))
	foreground := []rune(view.Message)
	background := []rune(view.Background)

	status := make([]rune, screenWidth)
	copy(status, head)
	sep := []rune(" > ")
	copy(status[len(head):len(head)+len(sep)], sep)
	if len(foreground) > 0 {
		copy(status[len(head)+len(sep):], foreground) // leave 1 space between head and foreground
	}
	if len(background) > 0 {
		background = append(sep, background...)
		copy(status[len(status)-len(background):], background)
	}

	for col, ch := range status {
		s.SetContent(col, screenHeight-1, ch, nil, statusStyle)
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

	e, flush, close, err := makeEditor(ctx, inputFilename, logFilename, width, height, nil)
	if err != nil {
		return err
	}
	defer close()

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
	// manual flush loop
	go func() {
		ticker := time.NewTicker(feature.LOG_FLUSH_INTERVAL_S * time.Second)
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
				flush()
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
func fileNonEmpty(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.Size() > 0
}

func handleEditorKey(e editor.Editor, event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyRune:
		e.Type(event.Rune())
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

	default:
	}
}
