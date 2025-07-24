package flag

import "os"

func Debug() bool {
	return len(os.Getenv("DEBUG")) > 0
}
