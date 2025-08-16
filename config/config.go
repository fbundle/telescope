package config

import (
	"telescope/util/atomic_util"
	"time"
)

const VERSION = "0.1.8a"

const HELP = `
Usage: "telescope [option] file [logfile]"
Options:
  -h --help           show help
  -v --version        get version
  -r --replay         replay the edited file 
  -l --log_writer            print the human readable log_writer format
  -i --insert         open with INSERT mode
  -c --command        open with NORMAL/COMMAND/VISUAL/INSERT mode
     --unsafe         open with UNSAFE mode

Keyboard Shortcuts:
  Ctrl+C              exit
  Ctrl+S              flush log_writer (autosave is always on, so this is not necessary)
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
  / :s :search                search
  :regex         search with regex
  : :g :goto          goto line
  :w :write         write into file
  :q :quit          quit
`

const (
	HUMAN_READABLE_SERIALIZER = 0
	BINARY_SERIALIZER         = 1
)

type Config struct {
	DEBUG                      bool
	VERSION                    string
	HELP                       string
	LOG_AUTOFLUSH_INTERVAL     time.Duration
	LOADING_PROGRESS_INTERVAL  time.Duration
	SERIALIZER_VERSION         uint64
	INITIAL_SERIALIZER_VERSION uint64
	MAXSIZE_HISTORY_STACK      int
	VIEW_CHANNEL_SIZE          int
	MAX_SEACH_TIME             time.Duration
	TAB_SIZE                   int
	LOG_DIR                    string
	TMP_DIR                    string
	SCROLL_SPEED               int
	LOAD_ESCAPE_INTERVAL       time.Duration
}

var config *atomic_util.Once[Config] = atomic_util.NewOnce[Config]()

func Load() Config {
	return config.LoadOrStore(loadConfig)
}
