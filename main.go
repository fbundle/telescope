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

func draw(s tcell.Screen, view editor.View) {
	s.Clear()
	screenWidth, screenHeight := s.Size()
	// Draw the title bar in reverse color
	title := []rune(fmt.Sprintf("Telescope: Editting \"%s\" - Ctrl+S to save Ctrl+C to quit", filenameOut))
	for len(title) < screenWidth {
		title = append(title, ' ')
	}
	for col, ch := range title {
		s.SetContent(col, 0, ch, nil, tcell.StyleDefault.Reverse(true))
	}
	// Draw editor content 1 row from the top
	for row, line := range view.Data {
		for col, ch := range line {
			s.SetContent(col, row+1, ch, nil, tcell.StyleDefault)
		}
	}

	// Draw status bar at the bottom
	status := []rune(view.Status)
	for len(status) < screenWidth {
		status = append(status, ' ')
	}
	for col, ch := range status {
		s.SetContent(col, screenHeight-1, ch, nil, tcell.StyleDefault.Reverse(true))
	}

	// Draw cursor 1 row from the top
	s.ShowCursor(view.Cursor.Col, view.Cursor.Row+1)
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
	backendEditor = editor.NewEditor(height-2, width, filenameIn, filenameOut)

	// initial draw
	draw(s, backendEditor.Render())

	// update loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-backendEditor.Update():
				draw(s, backendEditor.Render())
			}
		}
	}()

	// even loop
	running := true
	for running {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			handleKey(ev)
			draw(s, backendEditor.Render())
			if tcell.KeyCtrlC == ev.Key() {
				running = false
			}
		case *tcell.EventResize:
			s.Sync()
			width, height = s.Size()
			backendEditor.Resize(height-2, width)
			draw(s, backendEditor.Render())
		default:
			// nothing
		}
	}
}
