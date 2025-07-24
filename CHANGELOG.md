- very basic text editor
- added basic keyboard support: `Ctrl+C,Ctrl+S`, `Enter`, `Backspace`, `Delete`, `Left,Right,Up,Down,Home,End,PgUp,PgDn`
- first working code - file are loaded lazily
- improved performance with mmap
- added option for read only
- added loading status - basically execute draw every second
- [experimental] added parallel indexing - use environment variable to enable `PARALLEL_INDEXING=1`
- fixed for too short characters - render give the whole row not according to window width anymore
- change frontend to tcell - much better
- improved structure, improve frontend

TODO - add saving progress - possibly requires a bit of abstraction for unconditionally update
TODO - add vim-like command at status bar
TODO - fix bug with reading and writing to the same file
TODO - add journaling, crash recovery
TODO - add undo,redo
TODO - add search
TODO - add goto line number
TODO - change to journalling mode, there is no save, just the original file and change history and a separate command to combine into a new file
