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
	DEBUG_IO_INTERVAL_MS         time.Duration
	LOG_FLUSH_INTERVAL_S         time.Duration
	LOADING_PROGRESS_INTERVAL_MS time.Duration
	SERIALIZER_VERSION           uint64
	INITIAL_SERIALIZER_VERSION   uint64
	MAXSIZE_HISTORY              int
}

// TODO - export these into environment variables

func Load() Config {
	return Config{
		DEBUG_IO_INTERVAL_MS:         100 * time.Millisecond,
		LOG_FLUSH_INTERVAL_S:         60 * time.Second,
		LOADING_PROGRESS_INTERVAL_MS: 100 * time.Millisecond,
		SERIALIZER_VERSION:           BINARY_SERIALIZER,
		INITIAL_SERIALIZER_VERSION:   HUMAN_READABLE_SERIALIZER,
		MAXSIZE_HISTORY:              1024,
	}
}

func Debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}
