package app

import (
	"context"
	"fmt"
	"os"
	"telescope/editor"
	"telescope/log"
)

func RunReplay(inputFilename string, logFilename string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _ = fmt.Fprintf(os.Stderr, "loading input file %s\n", inputFilename)
	loadCtx, loadCancel := context.WithCancel(ctx)
	e, err := editor.NewEditor(
		ctx,
		20, 20,
		inputFilename, "",
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
	_, _ = fmt.Fprintf(os.Stderr, "loading log file %s\n", logFilename)

	err = log.Read(logFilename, func(entry log.Entry) bool {
		switch entry.Command {
		case log.CommandEnter:
			e.Jump(int(entry.CursorRow), int(entry.CursorCol))
			e.Enter()
		case log.CommandBackspace:
			e.Jump(int(entry.CursorRow), int(entry.CursorCol))
			e.Backspace()
		case log.CommandDelete:
			e.Jump(int(entry.CursorRow), int(entry.CursorCol))
			e.Delete()
		case log.CommandType:
			e.Jump(int(entry.CursorRow), int(entry.CursorCol))
			e.Type(entry.Rune)
		default:
			panic("unknown command")
		}
		return true
	})
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stderr, "replaying file\n")
	for _, line := range e.Iter {
		_, _ = fmt.Fprintln(os.Stdout, string(line))
	}
	return nil
}
