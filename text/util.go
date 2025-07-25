package text

import (
	"context"
	"slices"
	"telescope/config"
	"telescope/exit"
	"time"

	"golang.org/x/exp/mmap"
)

func padNewLine(line []byte) []byte {
	if line[len(line)-1] == '\n' {
		return line
	}
	return append(slices.Clone(line), '\n')
}

func endOfLineSize(line []byte) int {
	if len(line) == 0 {
		exit.Write("empty line")
	}
	if line[len(line)-1] != '\n' {
		exit.Write("not end of line")
	}
	if len(line) >= 2 && line[len(line)-2] == '\r' {
		// for windows
		return 2
	}
	return 1
}

func indexFile(ctx context.Context, reader *mmap.ReaderAt, delim byte, update func(offset int, line []byte)) error {
	var offset int = 0

	for i := 0; i < reader.Len(); i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if reader.At(i) == delim {
			line := make([]byte, i+1-offset)
			_, _ = reader.ReadAt(line, int64(offset))
			if config.Debug() {
				time.Sleep(config.Load().DEBUG_IO_INTERVAL_MS * time.Millisecond)
			}
			update(offset, line)
			offset += len(line)
		}
	}
	if offset < reader.Len() { // file doesn't end with trailing new line
		line := make([]byte, reader.Len()-offset)
		_, err := reader.ReadAt(line, int64(offset))
		if err != nil {
			return err
		}
		if config.Debug() {
			time.Sleep(config.Load().DEBUG_IO_INTERVAL_MS * time.Millisecond)
		}
		update(offset, line)
	}
	return nil
}
