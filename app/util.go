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

func fileNonEmpty(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.Size() > 0
}
func makeEditor(ctx context.Context, inputFilename string, logFilename string, width int, height int) (editor.Editor, context.Context, func() error, func(), error) {
	closerList := make([]func() error, 0)
	closer := func() {
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
			closer()
			return nil, nil, nil, nil, err
		}
		closerList = append(closerList, inputMmapReader.Close)
		winName += " " + filepath.Base(inputFilename)
	}

	// log
	var logFile *os.File = nil
	var logWriter log.Writer = nil
	var flush func() error = func() error { return nil }
	if len(logFilename) > 0 {
		logFile, err = os.OpenFile(logFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			closer()
			return nil, nil, nil, nil, err
		}
		closerList = append(closerList, logFile.Close)
		writer := bufio.NewWriter(logFile)

		closerList = append(closerList, writer.Flush) // flush before closer

		logWriter, err = log.NewWriter(writer)
		if err != nil {
			closer()
			return nil, nil, nil, nil, err
		}
		flush = writer.Flush
	}

	// editor
	e, err := editor.NewEditor(
		height-1, width,
		logWriter,
	)
	if err != nil {
		closer()
		return nil, nil, nil, nil, err
	}

	// load input file
	loadCtx, err := e.Load(ctx, inputMmapReader)
	if err != nil {
		closer()
		return nil, nil, nil, nil, err
	}

	return e, loadCtx, flush, closer, err
}
func writeMessage(e editor.Editor, message string) {
	e.Status(func(status editor.Status) editor.Status {
		status.Message = message
		return status
	})
}
