package ui

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"telescope/core/editor"
	"telescope/core/insert_editor"
	"telescope/core/log_writer"
	"telescope/util/buffer"

	"golang.org/x/exp/mmap"
)

func makeInsertEditor(ctx context.Context, inputFilename string, logFilename string, width int, height int) (*insert_editor.Editor, context.Context, func() error, func(), error) {
	closerList := make([]func() error, 0)
	closer := func() {
		for i := len(closerList) - 1; i >= 0; i-- {
			closerList[i]()
		}
	}

	var err error
	// input text
	var inputBuffer buffer.Reader = nil
	var winName string = "telescope"
	if len(inputFilename) > 0 {
		inputMmapReader, err := mmap.Open(inputFilename)
		if err != nil {
			closer()
			return nil, nil, nil, nil, err
		}
		inputBuffer = inputMmapReader
		closerList = append(closerList, inputMmapReader.Close)
		winName += " " + filepath.Base(inputFilename)
	}

	// log_writer
	var logFile *os.File = nil
	var logWriter log_writer.Writer = nil
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

		logWriter, err = log_writer.NewWriter(writer)
		if err != nil {
			closer()
			return nil, nil, nil, nil, err
		}
		flush = writer.Flush
	}

	// insert_editor
	e, err := insert_editor.New(
		height-1, width,
	)
	if err != nil {
		closer()
		return nil, nil, nil, nil, err
	}
	e.Subscribe(func(entry editor.LogEntry) {
		_ = logWriter.Write(entry)
	})

	// load input file
	loadCtx, err := e.Load(ctx, inputBuffer)
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
