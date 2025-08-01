package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"telescope/ui"
	"telescope/util/side_channel"
)

const VERSION = "0.1.7b"

var HELP = `
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

func printHelp() {
	fmt.Printf("telescope version %s\n", VERSION)
	fmt.Println(HELP)
}

func printVersion() {
	fmt.Println(VERSION)
}

type programArgs struct {
	option         string
	firstFilename  string
	secondFilename string
}

func main() {
	args := getProgramArgs()
	if len(args.firstFilename) == 0 {
		printHelp()
		return
	}
	switch args.option {
	case "-h", "--help":
		printHelp()
		return
	case "-v", "--version":
		printVersion()
		return
	case "-r", "--replay":
		if err := ui.RunReplay(args.firstFilename, args.secondFilename); err != nil {
			log.Fatalln(err)
		}
	case "-l", "--log":
		err := ui.RunLog(args.firstFilename)
		if err != nil {
			side_channel.Panic(err)
		}
	case "-i", "--insert":
		if fileNonEmpty(args.secondFilename) {
			ok := promptYesNo(fmt.Sprintf("log file exists (%s), delete it?", args.secondFilename), false)
			if !ok {
				return
			}
			err := os.Remove(args.secondFilename)
			if err != nil {
				side_channel.Panic(err)
			}
		}
		err := ui.RunEditor(args.firstFilename, args.secondFilename, false)
		if err != nil {
			side_channel.Panic(err)
		}
	case "-c", "--command":
		if fileNonEmpty(args.secondFilename) {
			ok := promptYesNo(fmt.Sprintf("log file exists (%s), delete it?", args.secondFilename), false)
			if !ok {
				return
			}
			err := os.Remove(args.secondFilename)
			if err != nil {
				side_channel.Panic(err)
			}
		}
		err := ui.RunEditor(args.firstFilename, args.secondFilename, true)
		if err != nil {
			side_channel.Panic(err)
		}
	default:
		// by default - open with command mode
		if fileNonEmpty(args.secondFilename) {
			ok := promptYesNo(fmt.Sprintf("log file exists (%s), delete it?", args.secondFilename), false)
			if !ok {
				return
			}
			err := os.Remove(args.secondFilename)
			if err != nil {
				side_channel.Panic(err)
			}
		}
		err := ui.RunEditor(args.firstFilename, args.secondFilename, true)
		if err != nil {
			side_channel.Panic(err)
		}
	}
}

func fileNonEmpty(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.Size() > 0
}

func promptYesNo(prompt string, defaultOption bool) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		if defaultOption {
			fmt.Print(prompt + " [Y/n]: ")
		} else {
			fmt.Print(prompt + " [y/N]: ")
		}
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if len(input) == 0 {
			return defaultOption
		}
		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		} else {
			fmt.Println("Please enter y or n.")
		}
	}
}
func getDefaultLogFilename(inputFilename string) string {
	absPath, err := filepath.Abs(inputFilename)
	if err != nil {
		side_channel.Panic(err)
		return ""
	}
	tempDir := os.TempDir()
	destPath := filepath.Join(tempDir, "telescope_log", absPath)
	err = os.MkdirAll(filepath.Dir(destPath), 0o700)
	if err != nil {
		side_channel.Panic(err)
		return ""
	}

	return destPath
}

func consume(args []string) ([]string, string) {
	if len(args) == 0 {
		return []string{}, ""
	}
	return args[1:], args[0]
}
func peek(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
func getProgramArgs() programArgs {
	args := os.Args[1:]
	pargs := programArgs{}

	if head := peek(args); len(head) > 0 && head[0] == '-' {
		pargs.option = head
		args, _ = consume(args)
	}
	args, pargs.firstFilename = consume(args)
	args, pargs.secondFilename = consume(args)
	if len(pargs.secondFilename) == 0 {
		pargs.secondFilename = getDefaultLogFilename(pargs.firstFilename)
	}
	return pargs
}
