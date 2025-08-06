package multimode_editor

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"telescope/config"
)

func safeWriteFile(filename string, iter func(f func(i int, val []rune) bool)) error {
	copyFileContents := func(src string, dst string) error {
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
	overwriteFile := func(dstFilename string, srcFilename string) error {
		// save mode
		dstInfo, err := os.Stat(dstFilename)
		if err != nil {
			return err
		}
		dstMode := dstInfo.Mode()

		// save file
		err = os.Rename(srcFilename, dstFilename)
		if err != nil {
			// try copy file content
			err = copyFileContents(srcFilename, dstFilename)
			if err != nil {
				return err
			}
		}

		// restore mode
		err = os.Chmod(dstFilename, dstMode)
		if err != nil {
			return err
		}
		return nil
	}

	writeFile := func(filename string, iter func(f func(i int, val []rune) bool)) error {

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

	absPath, _ := filepath.Abs(filename)
	tmpFilename := filepath.Join(config.Load().TMP_DIR, absPath)
	err := os.MkdirAll(filepath.Dir(tmpFilename), 0o700)
	if err != nil {
		return err
	}

	defer os.Remove(tmpFilename) // remove tmp file at the end

	// write into tmp file
	err = writeFile(tmpFilename, iter)
	if err != nil {
		return err
	}

	// move tmp file into output file
	return overwriteFile(filename, tmpFilename)
}
