package feature

import "os"

func Debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}

func ParallelIndexing() bool {
	return len(os.Getenv("PARALLEL_INDEXING")) > 0
}
