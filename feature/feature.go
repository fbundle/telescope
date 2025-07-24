package feature

import "os"

func Debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}

func ParallelIndexing() bool {
	return len(os.Getenv("PARALLEL_INDEXING")) > 0
}

const (
	JOURNAL_INTERVAL_S   = 60
	DEBUG_IO_INTERVAL_MS = 100
)
