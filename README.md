# TELESCOPE

an extremely fast text editor

![screenshot](./screenshots/0_1_2.png)

## FEATURE SET

- basic text editor
    - navigate with `Left`, `Right`, `Up`, `Down`, `PgUp`, `PgDn`, `Home`, `End`
    - undo redo with `Ctrl+U`, `Ctrl+R`

- instant start-up, instant edit

- able to handle very large files, potentially even larger than system memory.

- able to recover from crash

- able to edit while still loading the file and exit without losing any progress

- [experimental - with known bugs] add vim-like command mode, search, goto line, etc.

  - the text editor is now has 3 states/modes: `VISUAL/COMMAND/INSERT`

  - the text editor starts with `VISUAL` mode, when user types `:` it will switch to `COMMAND` mode

  - in `COMMAND` mode, if user keeps typing, it will write into the command buffer, and when user press `ENTER`, the command is executed

  - command `:i` or `:insert` will change the mode to INSERT

  - command `:s <pattern>` or `:search <pattern>` will do a search

  - command `:g <line_number>` or `:goto <line_number>` will go to a certain line

  - command `:w <file_name>` or `:write <file_name>` write write the current content into a new file (must be a new file)

  - after a command is executed, except `:i` or `:insert`, command buffer will be deleted and the editor will go back to `VISUAL` mode

  - in `INSERT` mode, user will edit the file as usual

  - in any mode, press `ESCAPE` will delete the command buffer and go back to `VISUAL` mode

## INTERACTION WITH FILE SYSTEM

0. use `telescope -h` for help

1. when user opens a file using `telescope dir/file`, the program will create a log file (journal file) at `dir/.file.log` (if the directory is not writable, user can specify the log file by `telescope dir/file logdir/logfile`).

2. when user edit the file, every action will be written to log file.

3. when exit the program the log file is preserved to export

4. user use `telescope -r dir/file` to apply the log together with original file to make a new file. the program will write the output to stdout, so user should be `telescope -r dir/file 1> outputfile`

