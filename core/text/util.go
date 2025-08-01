package text

import (
	"context"
	"slices"
	"telescope/util/buffer"
	"telescope/util/side_channel"
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
func LoadFile(ctx context.Context, reader buffer.Buffer, update func(Line, int)) error {
	return indexFile(ctx, reader, '\n', func(offset int, line []byte) {
		l := makeLineFromFile(offset)
		update(l, len(line))
	})
}

func indexFile(ctx context.Context, reader buffer.Buffer, delim byte, update func(offset int, line []byte)) error {
	var offset int = 0

	var line []byte

	for i := 0; i < reader.Len(); i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		b := reader.At(i)
		if b != delim {
			line = append(line, b)
			continue
		}
		update(offset, line)
		offset, line = offset+len(line)+1, nil
	}
	if len(line) > 0 {
		update(offset, line)
	}
	return nil
}
