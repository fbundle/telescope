package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"telescope/editor"

	"github.com/gdamore/tcell/v2"
)

const VERSION = "0.1.2"

var filenameIn, filenameOut string
var backendEditor editor.Editor

var statusStyle = tcell.StyleDefault.
	Background(tcell.ColorLightGray).
	Foreground(tcell.ColorBlack)

var textStyle = tcell.StyleDefault

func draw(s tcell.Screen, view editor.View) {
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
	if len(head) > 0 {
		copy(status, head)
	}
	if len(foreground) > 0 {
		copy(status[len(head)+1:], foreground) // leave 1 space between head and foreground
	}
	if len(background) > 0 {
		copy(status[len(status)-len(background):], background)
	}

	for col, ch := range status {
		s.SetContent(col, screenHeight-1, ch, nil, statusStyle)
	}

	s.Show()
}

func handleKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyCtrlC:
		// do nothing
	case tcell.KeyCtrlS:
		backendEditor.Save()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		backendEditor.Backspace()
	case tcell.KeyDelete:
		backendEditor.Delete()
	case tcell.KeyRight:
		backendEditor.MoveRight()
	case tcell.KeyLeft:
		backendEditor.MoveLeft()
	case tcell.KeyUp:
		backendEditor.MoveUp()
	case tcell.KeyDown:
		backendEditor.MoveDown()
	case tcell.KeyHome:
		backendEditor.MoveHome()
	case tcell.KeyEnd:
		backendEditor.MoveEnd()
	case tcell.KeyPgUp:
		backendEditor.MovePageUp()
	case tcell.KeyPgDn:
		backendEditor.MovePageDown()
	case tcell.KeyEnter:
		backendEditor.Enter()
	case tcell.KeyEsc:
		backendEditor.Escape()
	case tcell.KeyTab:
		backendEditor.Tabular()
	case tcell.KeyRune:
		backendEditor.Type(ev.Rune())
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: telescope <input file> <output file>")
		os.Exit(1)
	}
	if len(os.Args) < 3 {
		filenameIn, filenameOut = os.Args[1], ""
		if filenameIn == "--version" {
			fmt.Printf("telescope version %s\n", VERSION)
			os.Exit(0)
		}
	} else {
		filenameIn, filenameOut = os.Args[1], os.Args[2]
	}

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("cannot create screen: %v", err)
	}
	if err = s.Init(); err != nil {
		log.Fatalf("cannot initialize screen: %v", err)
	}
	defer s.Fini()

	width, height := s.Size()
	backendEditor, err = editor.NewEditor(height-1, width, filenameIn, filenameOut)
	if err != nil {
		panic(err)
	}

	// draw loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case view := <-backendEditor.Update():
				draw(s, view)
			}
		}
	}()

	// event loop
	running := true
	for running {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			handleKey(ev)
			if tcell.KeyCtrlC == ev.Key() {
				running = false
			}
		case *tcell.EventResize:
			s.Sync()
			width, height = s.Size()
			backendEditor.Resize(height-1, width)
		default:
			// nothing
		}
	}
}
