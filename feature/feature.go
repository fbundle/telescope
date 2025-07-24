package feature

import "os"

func Debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}

func ParallelIndexing() bool {
	return len(os.Getenv("PARALLEL_INDEXING")) > 0
}

func Journaling() bool {
	return len(os.Getenv("JOURNALING")) > 0
}

const (
	JOURNAL_INTERVAL_S           = 60
	DEBUG_IO_INTERVAL_MS         = 100
	LOADING_PROGRESS_INTERVAL_MS = 100
)
