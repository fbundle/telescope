package text

import (
	"context"
	"telescope/config"
	"telescope/util/buffer"
)

func IndexFile(ctx context.Context, reader buffer.Reader, update func(offset int, line []byte)) error {
	var offset int = 0
	var line []byte = nil

	for i := 0; i < reader.Len(); i++ {
		if i%config.Load().LOAD_ESCAPE_INTERVAL_BYTES == 0 {
			// check every 10MB
			select {
			case <-ctx.Done():
				return nil
			default:
			}
		}

		b := reader.At(i)
		line = append(line, b)
		if line[len(line)-1] == delim {
			update(offset, line) // line with delim
			offset, line = i+1, nil
		}
	}
	if len(line) > 0 {
		update(offset, line)
	}
	return nil
}
