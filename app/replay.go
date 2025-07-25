package app

import (
	"context"
	"fmt"
	"os"
	"telescope/log"
)

func RunReplay(inputFilename string, logFilename string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _ = fmt.Fprintf(os.Stderr, "loading input file %s\n", inputFilename)

	// make editor without log
	e, loadCtx, flush, close, err := makeEditor(ctx, inputFilename, "", 20, 20)
	if err != nil {
		return err
	}
	defer close()
	_ = flush // do nothing with flush

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case view := <-e.Update(): // consume view
				if len(view.Background) > 0 {
					_, _ = fmt.Fprintln(os.Stderr, view.Background)
				}
			}
		}
	}()

	// wait for loading
	<-loadCtx.Done()

	_, _ = fmt.Fprintf(os.Stderr, "loading log file %s\n", logFilename)

	err = log.Read(logFilename, func(entry log.Entry) bool {
		e.Apply(entry)
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
