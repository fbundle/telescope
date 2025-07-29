package text

import (
	"context"
	"slices"
	"telescope/config"
	"telescope/core/bytes"
	"telescope/util/side_channel"
	"time"
)

func padNewLine(line []byte) []byte {
	if line[len(line)-1] == '\n' {
		return line
	}
	return append(slices.Clone(line), '\n')
}

func endOfLineSize(line []byte) int {
	if len(line) == 0 {
		side_channel.Panic("empty line")
	}
	if line[len(line)-1] != '\n' {
		side_channel.Panic("not end of line")
	}
	if len(line) >= 2 && line[len(line)-2] == '\r' {
		// windows line ends with \r\n
		return 2
	}
	// linux/macos line ends with \n
	return 1
}

func indexFile(ctx context.Context, reader bytes.Array, delim byte, update func(offset int, line []byte)) error {
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
				time.Sleep(config.Load().DEBUG_IO_DELAY)
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
			time.Sleep(config.Load().DEBUG_IO_DELAY)
		}
		update(offset, line)
	}
	return nil
}
