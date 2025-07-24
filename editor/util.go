package editor

import (
	"log"
	"os"
)

func insertToSlice[T any](l []T, i int, v T) []T {
	var zero T
	l = append(l, zero)
	copy(l[i+1:], l[i:])
	l[i] = v
	return l
}

func deleteFromSlice[T any](l []T, i int) []T {
	copy(l[i:], l[i+1:])
	return l[:len(l)-1]
}
func concatSlices[T any](ls ...[]T) []T {
	l := make([]T, 0)
	for _, s := range ls {
		l = append(l, s...)
	}
	return l
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func fileSize(filename string) int {
	info, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	size := info.Size() // in bytes
	return int(size)
}
