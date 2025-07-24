package app

import (
	"context"
	"fmt"
	"os"
	"telescope/editor"
	"telescope/journal"
)

func RunReplay(inputFilename string, journalFilename string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _ = fmt.Fprintf(os.Stderr, "loading input file %s\n", inputFilename)
	loadCtx, loadCancel := context.WithCancel(ctx)
	e, err := editor.NewEditor(
		ctx,
		20, 20,
		inputFilename, "", "",
		loadCancel,
	)
	if err != nil {
		return err
	}
	go func() {
		for range e.Update() {
			// consume view
		}
	}()
	<-loadCtx.Done()
	_, _ = fmt.Fprintf(os.Stderr, "loading journal file %s\n", journalFilename)

	err = journal.Read(journalFilename, func(entry journal.Entry) bool {
		switch entry.Command {
		case journal.CommandEnter:
			e.Jump(entry.CursorRow, entry.CursorCol)
			e.Enter()
		case journal.CommandBackspace:
			e.Jump(entry.CursorRow, entry.CursorCol)
			e.Backspace()
		case journal.CommandDelete:
			e.Jump(entry.CursorRow, entry.CursorCol)
			e.Delete()
		case journal.CommandType:
			e.Jump(entry.CursorRow, entry.CursorCol)
			e.Type(entry.Rune)
		default:
			panic("unknown command")
		}
		return true
	})
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stderr, "replaying file")
	for _, line := range e.Iter {
		_, _ = fmt.Fprintf(os.Stdout, string(line)+"\n")
	}
	return nil
}
