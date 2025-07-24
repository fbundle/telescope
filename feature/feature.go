package feature

import "os"

func Debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}

func ParallelIndexing() bool {
	return len(os.Getenv("PARALLEL_INDEXING")) > 0
}

func DisableJournal() bool {
	return len(os.Getenv("DISABLE_JOURNAL")) > 0
}

const (
	JOURNAL_INTERVAL_S           = 1
	DEBUG_IO_INTERVAL_MS         = 100
	LOADING_PROGRESS_INTERVAL_MS = 100
)
