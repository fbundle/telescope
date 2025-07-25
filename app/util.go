package app

import (
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

		logWriter, err = log.NewWriter(logFile)
		if err != nil {
			close()
			return nil, nil, nil, err
		}
		flush = logFile.Sync
	}

	// editor
	e, err := editor.NewEditor(
		ctx,
		winName,
		height-1, width,
		inputMmapReader, logWriter,
		loadDone,
	)
	if err != nil {
		close()
		return nil, nil, nil, err
	}
	return e, flush, close, err
}
