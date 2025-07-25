package feature

import (
	"os"
)

func Debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}

func ParallelIndexing() bool {
	return len(os.Getenv("PARALLEL_INDEXING")) > 0
}

const (
	DEBUG_IO_INTERVAL_MS         = 100
	LOADING_PROGRESS_INTERVAL_MS = 100
	DEFAULT_SERIALIZER_VERSION   = 1 // default for writer
	INITIAL_SERIALIZER_VERSION   = 0 // initial to read
)
