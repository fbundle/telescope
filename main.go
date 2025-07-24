package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"telescope/editor"

	"github.com/jroimartin/gocui"
)

const (
	VERSION = "0.1.1"
)

const (
	statusViewName = "status"
	editorViewName = "editor"
)

var filenameIn, filenameOut string
var backendEditor editor.Editor = nil

type customEditor struct {
}

func (e *customEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		backendEditor.Type(ch)
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		backendEditor.Backspace()
	case key == gocui.KeyDelete:
		backendEditor.Delete()
	case key == gocui.KeyArrowRight:
		backendEditor.MoveRight()
	case key == gocui.KeyArrowLeft:
		backendEditor.MoveLeft()
	case key == gocui.KeyArrowUp:
		backendEditor.MoveUp()
	case key == gocui.KeyArrowDown:
		backendEditor.MoveDown()
	case key == gocui.KeySpace:
		backendEditor.Type(' ')
	case key == gocui.KeyCtrlS:
		backendEditor.Save()
	case key == gocui.KeyHome:
		backendEditor.MoveHome()
	case key == gocui.KeyEnd:
		backendEditor.MoveEnd()
	case key == gocui.KeyPgup:
		backendEditor.MovePageUp()
	case key == gocui.KeyPgdn:
		backendEditor.MovePageDown()
	case key == gocui.KeyEnter:
		backendEditor.Enter()
	}
}
func draw(statusView *gocui.View, editorView *gocui.View, view editor.View) error {
	statusView.Clear()
	_, err := statusView.Write([]byte(view.Status + "\n"))
	if err != nil {
		return err
	}

	editorView.Clear()
	for _, line := range view.Data {
		_, err = editorView.Write([]byte(string(line) + "\n"))
		if err != nil {
			return err
		}
	}
	err = editorView.SetCursor(view.Cursor.Col, view.Cursor.Row)
	if err != nil {
		fmt.Println(view.Cursor)
		return err
	}
	return nil
}

func layout(g *gocui.Gui) error {
	termWidth, termHeight := g.Size()

	// update editor
	editorHeight, editorWidth := (termHeight-3)-2, termWidth-2
	backendEditor.Resize(editorHeight, editorWidth)

	// setup status view
	statusView, err := g.SetView(statusViewName, 0, termHeight-3, termWidth-1, termHeight-1)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}

	// setup editor view
	editorView, err := g.SetView(editorViewName, 0, 0, termWidth-1, termHeight-4)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return err
	}
	editorView.Editor = &customEditor{}
	editorView.Editable = true
	editorView.Title = fmt.Sprintf("Telescope: Editting \"%s\" - Ctrl+S to save Ctrl+C to quit", filenameOut)

	_, err = g.SetCurrentView(editorViewName)
	if err != nil {
		return err
	}
	view := backendEditor.Render()
	return draw(statusView, editorView, view)
}

func drawLoop(ctx context.Context, g *gocui.Gui) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-backendEditor.Update():
			g.Update(layout)
		}
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
	backendEditor = editor.NewEditor(24, 80, filenameIn, filenameOut)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.Cursor = true
	g.SetManagerFunc(layout)

	ctx, cancel := context.WithCancel(context.Background())
	go drawLoop(ctx, g)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		cancel()
		return gocui.ErrQuit
	}); err != nil {
		log.Fatal(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Fatal(err)
	}

}
