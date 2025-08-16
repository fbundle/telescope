package config

import (
	"os"
	"path/filepath"
	"strconv"
	"telescope/util/side_channel"
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

func loadConfig() Config {
	tempDir := os.TempDir()
	side_channel.WriteLn("temp dir:", tempDir)
	defaultLogDir := filepath.Join(tempDir, "telescope", "log")
	defaultTmpDir := filepath.Join(tempDir, "telescope", "tmp")
	return Config{
		DEBUG:                      loadConfVarBool("DEBUG", false),
		VERSION:                    VERSION,
		HELP:                       HELP,
		LOG_AUTOFLUSH_INTERVAL:     LoadConfVarDuration("LOG_AUTOFLUSH_INTERVAL", 5*time.Second),
		LOADING_PROGRESS_INTERVAL:  LoadConfVarDuration("LOADING_PROGRESS_INTERVAL", 100*time.Millisecond),
		SERIALIZER_VERSION:         HUMAN_READABLE_SERIALIZER,
		INITIAL_SERIALIZER_VERSION: HUMAN_READABLE_SERIALIZER,
		MAXSIZE_HISTORY_STACK:      loadConfVarInt("MAXSIZE_HISTORY_STACK", 128),
		VIEW_CHANNEL_SIZE:          loadConfVarInt("VIEW_CHANNEL_SIZE", 16),
		MAX_SEACH_TIME:             LoadConfVarDuration("MAX_SEACH_TIME", 5*time.Second),
		TAB_SIZE:                   loadConfVarInt("TAB_SIZE", 2),
		LOG_DIR:                    loadConfVarString("LOG_DIR", defaultLogDir),
		TMP_DIR:                    loadConfVarString("TMP_DIR", defaultTmpDir),
		SCROLL_SPEED:               loadConfVarInt("SCROLL_SPEED", 3),
		LOAD_ESCAPE_INTERVAL:       LoadConfVarDuration("LOAD_ESCAPE_INTERVAL", 100*time.Millisecond),
	}

}
