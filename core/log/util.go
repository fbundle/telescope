package log

import (
	"encoding/binary"
	"errors"
	"io"
)

func uint64ToBytes(x uint64) []byte {
	b := make([]byte, 8) // 8 bytes
	binary.LittleEndian.PutUint64(b, x)
	return b
}

func bytesToUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

func runeToBytes(x rune) []byte {
	b := make([]byte, 4)                        // 4 bytes
	binary.LittleEndian.PutUint32(b, uint32(x)) // reinterpret rune int32 as uint32
	return b
}

func bytesToRune(b []byte) rune {
	return rune(binary.LittleEndian.Uint32(b))
}

func lengthPrefixWrite(w io.Writer, b []byte) error {
	lb := uint64ToBytes(uint64(len(b)))
	buf := append(lb, b...)

	_, err := w.Write(buf)
	return err
}

func lengthPrefixRead(r io.Reader) ([]byte, error) {
	lb := make([]byte, 8)
	_, err := r.Read(lb)
	if err != nil {
		return nil, err
	}
	l := bytesToUint64(lb)

	b := make([]byte, l)
	_, err = r.Read(b)
	if err != nil {
		return nil, errors.New("incomplete read")
	}
	return b, nil
}
