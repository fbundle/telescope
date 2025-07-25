# FEATURE

- very basic text editor
- added basic keyboard support: `Ctrl+C,Ctrl+S`, `Enter`, `Backspace`, `Delete`, `Left,Right,Up,Down,Home,End,PgUp,PgDn`
- first working code - file are loaded lazily
- improved performance with mmap
- added option for read only
- added loading status - basically execute draw every second
- [experimental] added parallel indexing, use `PARALLEL_INDEXING=1` to enable
- fixed for too short characters - render give the whole row not according to window width anymore
- change frontend to tcell - much better
- improved structure, improve frontend
- added journal file (file is saved every 10 seconds or when buffer is full), use `DISABLE_JOURNAL=1` to disable
- add journal file replay, remove save function completely. if user wants to save a file, they has to use `telescope -r <input_file>`, the content will be written to stdout. `DISABLE_JOURNAL=1` no longer works
- rename journal to log
- fix a critical bug, printing using fmt.Fprintf doesn't escape `%` in code
- log is now use length-prefixed format so that the same file can be decoded using multiple format
- add function to print out human readable log format
- fixed exit condition - bufio.Writer is used, log will be written at the end of the program or every `LOG_FLUSH_INTERVAL_S=60` seconds
- fixed a bug that the UI keeps reading `mmap.ReaderAt` while is it already closed - just return `[]rune{}` if the file is closed

# TODO

- add vim-like command at status bar
- add undo,redo
- add search
- add goto line number
