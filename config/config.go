package config

import (
	"os"
	"time"
)

const VERSION = "0.1.7c"

const HELP = `
Usage: "telescope [option] file [logfile]"
Options:
  -h --help           show help
  -v --version        get version
  -r --replay         replay the edited file 
  -l --log            print the human readable log format
  -i --insert         open with INSERT mode
  -c --command        open with NORMAL/COMMAND/VISUAL/INSERT mode

Keyboard Shortcuts:
  Ctrl+C              exit
  Ctrl+S              flush log (autosave is always on, so this is not necessary)
  Ctrl+U              undo
  Ctrl+R              redo

NORMAL/COMMAND/VISUAL/INSERT mode:
  in NORMAL mode:
    i                 enter INSERT mode
    :                 enter COMMAND mode
    V                 enter VISUAL mode
    p                 paste from clipboard
  in COMMAND mode:
    ENTER             execute command
    ESCAPE            delete command buffer and enter NORMAL mode
  in INSERT mode:
    ESCAPE            enter NORMAL mode
  in VISUAL mode:
    up,dn,pgup,pgdn   move cursor and selector
    d                 cut into clipboard
    y                 copy into clipboard
    ESCAPE            enter NORMAL mode

Commands:
  :i :insert        enter INSERT mode
  :s :search        search
     :regex         search with regex
  :g :goto          goto line
  :w :write         write into file
  :q :quit          quit
`

const (
	HUMAN_READABLE_SERIALIZER = 0
	BINARY_SERIALIZER         = 1
)

type Config struct {
	VERSION                    string
	HELP                       string
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
		VERSION:                   VERSION,
		HELP:                      HELP,
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
