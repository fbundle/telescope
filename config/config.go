package config

import (
	"os"
	"time"
)

type Config struct {
	DEBUG_IO_INTERVAL_MS         time.Duration
	LOG_FLUSH_INTERVAL_S         time.Duration
	LOADING_PROGRESS_INTERVAL_MS time.Duration
	SERIALIZER_VERSION           uint64
	INITIAL_SERIALIZER_VERSION   uint64
}

func Load() Config {
	return Config{
		DEBUG_IO_INTERVAL_MS:         100 * time.Millisecond,
		LOG_FLUSH_INTERVAL_S:         60 * time.Second,
		LOADING_PROGRESS_INTERVAL_MS: 100 * time.Millisecond,
		SERIALIZER_VERSION:           1,
		INITIAL_SERIALIZER_VERSION:   0,
	}
}

func Debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}
