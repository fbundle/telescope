package app

import (
	"context"
	"fmt"
	"telescope/editor"
	"telescope/journal"
)

func RunRecoverFromJournal(inputFilename string, journalFilename string, outputFilename string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fmt.Printf("loading input file %s\n", inputFilename)
	loadCtx, loadCancel := context.WithCancel(ctx)
	e, err := editor.NewEditor(
		ctx,
		20, 20,
		inputFilename, journalFilename, outputFilename,
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
	fmt.Printf("loading journal file %s\n", journalFilename)

	err = journal.Read(ctx, journalFilename, func(entry journal.Entry) {
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
	})
	if err != nil {
		return err
	}
	fmt.Printf("saving file %s\n", outputFilename)
	e.Save()
	fmt.Println("done")
	return nil
}
