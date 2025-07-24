package journal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"telescope/feature"
	"time"
)

func Read(filename string, apply func(e Entry) bool) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		line = bytes.TrimSpace(line)
		if len(line) > 0 {
			var e Entry
			if err := json.Unmarshal(line, &e); err != nil {
				return err
			}
			if feature.Debug() {
				time.Sleep(feature.DEBUG_IO_INTERVAL_MS * time.Millisecond)
			}
			ok := apply(e)
			if !ok {
				return nil
			}
		}
		if err == io.EOF {
			return nil
		}
	}
}
