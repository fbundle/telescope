package text

import (
	"iter"
	"telescope/util/buffer"
)

func IndexFile(reader buffer.Reader) iter.Seq2[int, int] {
	return func(yield func(index int, offset int) bool) {
		index, offset := 0, 0
		for i := 0; i < reader.Len(); i++ {
			b := reader.At(i)
			if b == delim {
				if !yield(index, offset) {
					return
				}
				index, offset = index+1, i+1
			}
		}
		if offset < reader.Len() {
			yield(index, offset)
		}
	}
}
