package text

import (
	"context"
	"telescope/config"
	"telescope/util/buffer"
)

func LoadFile(ctx context.Context, reader buffer.Reader, update func(Line)) error {
	return indexFile(ctx, reader, func(offset int, line []byte) {
		l := makeLineFromOffset(offset)
		update(l)
	})
}

func indexFile(ctx context.Context, reader buffer.Reader, update func(offset int, line []byte)) error {
	return indexFileFrom(ctx, reader, update, 0)
}

func LoadFileAfter(ctx context.Context, reader buffer.Reader, update func(Line), lastLine Line) error {
	beg := reader.Len()
	for i := int(lastLine.offset); i < reader.Len(); i++ {
		b := reader.At(i)
		if b == delim {
			beg = i + 1
			break
		}
	}

	return indexFileFrom(ctx, reader, func(offset int, line []byte) {
		l := makeLineFromOffset(offset)
		update(l)
	}, beg)
}

func indexFileFrom(ctx context.Context, reader buffer.Reader, update func(offset int, line []byte), beg int) error {
	var offset int = beg
	var line []byte = nil

	for i := beg; i < reader.Len(); i++ {
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
		if b == delim {
			update(offset, line) // line with delim
			offset, line = i+1, nil
		}
	}
	if len(line) > 0 {
		update(offset, line)
	}
	return nil
}
