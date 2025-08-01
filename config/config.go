package config

import (
	"os"
	"time"
)

const (
	HUMAN_READABLE_SERIALIZER = 0
	BINARY_SERIALIZER         = 1
)

type Config struct {
	DEBUG_IO_DELAY             time.Duration
	LOG_AUTOFLUSH_INTERVAL     time.Duration
	LOADING_PROGRESS_INTERVAL  time.Duration
	SERIALIZER_VERSION         uint64
	INITIAL_SERIALIZER_VERSION uint64
	MAXSIZE_HISTORY            int
	VIEW_CHANNEL_SIZE          int
	MAX_SEACH_TIME             time.Duration
	TAB_SIZE                   int
}

func Load() Config {
	// TODO - export these into environment variables
	return Config{
		DEBUG_IO_DELAY:            100 * time.Millisecond,
		LOG_AUTOFLUSH_INTERVAL:    60 * time.Second,
		LOADING_PROGRESS_INTERVAL: 100 * time.Millisecond,
		// SERIALIZER_VERSION:         BINARY_SERIALIZER,
		SERIALIZER_VERSION:         HUMAN_READABLE_SERIALIZER, // TODO - update serializer and enable binary version
		INITIAL_SERIALIZER_VERSION: HUMAN_READABLE_SERIALIZER,
		MAXSIZE_HISTORY:            1024,
		VIEW_CHANNEL_SIZE:          64,
		MAX_SEACH_TIME:             5 * time.Second,
		TAB_SIZE:                   2,
	}
}

func Debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}
