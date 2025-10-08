package file_util

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"telescope/config"
	"telescope/util/side_channel"
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

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}
func moveFile(dstFilename string, srcFilename string) error {
	// save mode
	var dstMode fs.FileMode = 0600
	if dstInfo, err := os.Stat(dstFilename); err == nil {
		dstMode = dstInfo.Mode()
	}

	// save file
	err := os.Rename(srcFilename, dstFilename)
	if err != nil {
		// try copy file content
		err = copyFile(srcFilename, dstFilename)
		if err != nil {
			return err
		}

		return os.Remove(srcFilename)
	}

	// restore mode
	err = os.Chmod(dstFilename, dstMode)
	if err != nil {
		return err
	}
	return nil
}

func SafeWriteFile(filename string, iter func(f func(i int, val []rune) bool)) error {

	absPath, _ := filepath.Abs(filename)
	tmpFilename := filepath.Join(config.Load().TMP_DIR, absPath)
	err := os.MkdirAll(filepath.Dir(tmpFilename), 0o700)
	if err != nil {
		side_channel.WriteLn(err)
		return err
	}

	defer os.Remove(tmpFilename) // remove tmp file at the end

	// write into tmp file
	err = writeFile(tmpFilename, iter)
	if err != nil {
		side_channel.WriteLn(err)
		return err
	}

	// move tmp file into output file
	err = moveFile(filename, tmpFilename)
	if err != nil {
		side_channel.WriteLn(err)
		return err
	}
	return nil
}
