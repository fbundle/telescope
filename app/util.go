package app

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"telescope/editor"
	"telescope/log"

	"golang.org/x/exp/mmap"
)

func makeEditor(ctx context.Context, inputFilename string, logFilename string, width int, height int, loadDone func()) (editor.Editor, func() error, func(), error) {
	closerList := make([]func() error, 0)
	close := func() {
		for i := len(closerList) - 1; i >= 0; i-- {
			closerList[i]()
		}
	}

	var err error
	// input text
	var inputMmapReader *mmap.ReaderAt = nil
	var winName string = "telescope"
	if fileNonEmpty(inputFilename) {
		inputMmapReader, err = mmap.Open(inputFilename)
		if err != nil {
			close()
			return nil, nil, nil, err
		}
		closerList = append(closerList, inputMmapReader.Close)
		winName += " " + filepath.Base(inputFilename)
	}

	// log
	var logFile *os.File = nil
	var logWriter log.Writer = nil
	var flush func() error = func() error { return nil }
	if len(logFilename) > 0 {
		logFile, err = os.OpenFile(logFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			close()
			return nil, nil, nil, err
		}
		closerList = append(closerList, logFile.Close)

		writer := bufio.NewWriter(logFile)

		logWriter, err = log.NewWriter(writer)
		if err != nil {
			close()
			return nil, nil, nil, err
		}
		flush = writer.Flush
	}

	loadCtx, loadCancel := context.WithCancel(ctx)
	// editor
	e, err := editor.NewEditor(
		ctx,
		winName,
		height-1, width,
		inputMmapReader, logWriter,
		func() {
			loadCancel()
			loadDone()
		},
	)
	if err != nil {
		loadCancel() // unable to create editor, just cancel the loadCtx anyway
		close()
		return nil, nil, nil, err
	}
	// waiting for loading to be done before closing
	closerList = append(closerList, func() error {
		<-loadCtx.Done()
		return nil
	})

	return e, flush, close, err
}
