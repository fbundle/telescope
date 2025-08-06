package command_editor

import (
	"bufio"
	"os"
)

func writeFile(filename string, iter func(f func(i int, val []rune) bool)) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, line := range iter {
		_, err = writer.WriteString(string(line) + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
