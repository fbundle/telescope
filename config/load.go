package config

import (
	"os"
	"runtime"
	"strconv"
	"telescope/util/side_channel"
	"time"
)

func loadConfVar[T any](varname string, parser func(string) (T, error), defaultVal T) T {
	if runtime.GOOS != "linux" {
		side_channel.WriteLn("disable conf_var from env since GOOS is not linux")
		return defaultVal
	}

	s, ok := os.LookupEnv("TELESCOPE_" + varname)
	if !ok {
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
