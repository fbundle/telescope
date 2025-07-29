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
		b, readErr := lengthPrefixRead(f)

		if readErr != nil && readErr != io.EOF {
			return readErr
		}

		s1, ok, processErr := func(s Serializer, b []byte) (Serializer, bool, error) {
			e, err := s.Unmarshal(b)
			if err != nil {
				return nil, true, err
			}
			switch e.Command {
			case CommandSetVersion:
				s1, err := GetSerializer(e.Version)
				return s1, true, err
			default:
				if config.Debug() {
					time.Sleep(config.Load().DEBUG_IO_DELAY)
				}
				return s, apply(e), nil
			}
		}(s, b)

		if processErr != nil {
			return processErr
		}
		if !ok || readErr == io.EOF {
			return nil
		}

		s = s1
	}
}
