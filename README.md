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

- [experimental] add vim-like command mode, search, goto line, etc.

  - command mode is enable when user starts with `telescope -c <filename>`

  - there are 3 states: `VISUAL/COMMAND/INSERT`, the initial state is `VISUAL`

  - in `VISUAL` mode, when user type `i`, the editor will go into `INSERT` mode

  - `INSERT` mode is the basic text editor

  - in `VISUAL` mode, when user type `:`, the editor will go into `COMMAND` mode

  - in `COMMAND` mode, typing will add into command buffer. when user presses `ENTER`, command will be executed

    - `:i` or `:insert` -> go into `INSERT` mode

    - `:s <pattern>` or `:search <pattern>` -> search for `<pattern>`

    - `:g <line_number>` or `:goto <line_number>` -> go into line `<line_number>`

    - `:w <file_name>` or `:write <file_name>` -> write the current file into `<file_name>`. note, writing into the same file is an undefined behavioriii
        ii undefined behavior (there is a bug in typing here), because the text editor doesn't load the input file into memory.

  - in any mode, pressing `ESCAPE` will clear command buffer and go to `VISUAL` mode
  

## INTERACTION WITH FILE SYSTEM

0. use `telescope -h` for help

1. when user opens a file using `telescope dir/file`, the program will create a log file (journal file) at `dir/.file.log` (if the directory is not writable, user can specify the log file by `telescope dir/file logdir/logfile`).

2. when user edit the file, every action will be written to log file.

3. when exit the program the log file is preserved to export

4. user use `telescope -r dir/file` to apply the log together with original file to make a new file. the program will write the output to stdout, so user should be `telescope -r dir/file 1> outputfile`

