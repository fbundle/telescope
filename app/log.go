package app

import (
	"fmt"
	"os"
	"strings"
	"telescope/config"
	"telescope/log"
)

func RunLog(logFilename string) error {
	s, err := log.GetSerializer(config.Load().INITIAL_SERIALIZER_VERSION)
	if err != nil {
		return err
	}

	log.Read(logFilename, func(e log.Entry) bool {
		var b []byte
		b, err = s.Marshal(e)
		if err != nil {
			return false
		}
		_, err = fmt.Fprintln(os.Stdout, strings.TrimSpace(string(b)))
		if err != nil {
			return false
		}
		return true
	})
	return err

}
