package log

import (
	"io"
	"os"
	"telescope/config"
	"time"
)

func Read(filename string, apply func(e Entry) bool) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	s, err := GetSerializer(config.Load().INITIAL_SERIALIZER_VERSION)
	if err != nil {
		return err
	}

	for {
		line, readErr := lengthPrefixRead(f)

		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		// process line, update serializer
		s1, ok, processErr := func(s Serializer, line []byte) (Serializer, bool, error) {
			if len(line) == 0 {
				return s, true, nil
			}
			e, err := s.Unmarshal(line)
			if err != nil {
				return nil, false, err
			}
			if e.Command == CommandSetVersion {
				// when log entry is a set_version, change the version of serializer
				s, err = GetSerializer(e.Version)
				return s, true, err
			}
			if config.Debug() {
				time.Sleep(config.Load().DEBUG_IO_DELAY)
			}
			return s, apply(e), nil
		}(s, line)

		//
		if processErr != nil || !ok || readErr == io.EOF {
			return err
		}

		s = s1
	}
}
