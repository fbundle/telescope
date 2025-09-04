package file_util

import (
	"os"
	"path/filepath"
)

func NonEmpty(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	if !info.Mode().IsRegular() {
		return false
	}
	return info.Size() > 0
}

func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	dirname := filepath.Dir(filename)
	if err := os.MkdirAll(dirname, 0700); err != nil {
		return err
	}
	return os.WriteFile(filename, data, perm)
}
