package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
	"strings"
	"telescope/editor"
	"telescope/feature"
	"telescope/journal"
)

const VERSION = "0.1.3"

var filenameTextIn, filenameTextOut string
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

func printHelp() {
	printVersion()
	help := `
Usage: telescope [option] <input_file> <output_file>
Option:
  -h --help	: show help
  -v --version	: get version
	`
	fmt.Println(help)
}

func printVersion() {
	fmt.Printf("telescope_extra version %s\n", VERSION)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		return
	}
	if os.Args[1] == "-v" || os.Args[1] == "--version" {
		printVersion()
		return
	}

	// text editor
	if len(os.Args) < 3 {
		filenameTextIn, filenameTextOut = os.Args[1], ""
	} else {
		filenameTextIn, filenameTextOut = os.Args[1], os.Args[2]
	}

	if !feature.DisableJournal() {
		journalFilename := journal.GetJournalFilename(filenameTextIn)
		if fileExists(journalFilename) && fileSize(journalFilename) > 0 {
			ok := promptYesNo(fmt.Sprintf("journal file exists (%s), delete it?", journalFilename), false)
			if !ok {
				return
			}
			err := os.Remove(journalFilename)
			if err != nil {
				panic(err)
			}
		}
	}

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("cannot create screen: %v", err)
	}
	if err = s.Init(); err != nil {
		log.Fatalf("cannot initialize screen: %v", err)
	}
	defer s.Fini()

	// draw loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	width, height := s.Size()
	backendEditor, err = editor.NewEditor(ctx, height-1, width, filenameTextIn, filenameTextOut, nil)
	if err != nil {
		panic(err)
	}
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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func fileSize(filename string) int {
	info, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	size := info.Size() // in bytes
	return int(size)
}
func promptYesNo(prompt string, defaultOption bool) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		if defaultOption {
			fmt.Print(prompt + " [Y/n]: ")
		} else {
			fmt.Print(prompt + " [y/N]: ")
		}
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if len(input) == 0 {
			return defaultOption
		}
		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		} else {
			fmt.Println("Please enter y or n.")
		}
	}
}
