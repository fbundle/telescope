package config

import (
	"os"
	"strconv"
	"time"
)

func loadConfVar[T any](varname string, parser func(string) (T, error), defaultVal T) T {
	s := os.Getenv(varname)
	if len(s) == 0 {
		return defaultVal
	}
	val, err := parser(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func loadConfVarInt(varname string, defaultVal int) int {
	return loadConfVar(varname, strconv.Atoi, defaultVal)
}

func LoadConfVarDuration(varname string, defaultVal time.Duration) time.Duration {
	return loadConfVar(varname, time.ParseDuration, defaultVal)
}

func loadConfVarBool(varname string, defaultVal bool) bool {
	return loadConfVar(varname, strconv.ParseBool, defaultVal)
}

func loadConfVarString(varname string, defaultVal string) string {
	return loadConfVar(varname, func(s string) (string, error) { return s, nil }, defaultVal)
}
