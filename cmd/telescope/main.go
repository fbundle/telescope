package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"telescope/config"
	"telescope/ui"
	"telescope/util/side_channel"
)

func printHelp() {
	fmt.Printf("telescope version %s\n", config.Load().VERSION)
	fmt.Println(config.Load().HELP)
}

func printVersion() {
	fmt.Printf("telescope version %s\n", config.Load().VERSION)
}

type programArgs struct {
	option         string
	firstFilename  string
	secondFilename string
}

func main() {
	args := getProgramArgs()
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
	case "-l", "--log_writer":
		err := ui.RunLog(args.firstFilename)
		if err != nil {
			side_channel.Panic(err)
		}
	case "-i", "--insert":
		if !promptDeleteLogFile(args) {
			return
		}
		err := ui.RunEditor(args.firstFilename, args.secondFilename, false)
		if err != nil {
			side_channel.Panic(err)
		}
	case "--unsafe":
		err := ui.RunEditor(args.firstFilename, "", true)
		if err != nil {
			side_channel.Panic(err)
		}
	default:
		// by default - open with command mode
		if !promptDeleteLogFile(args) {
			return
		}
		err := ui.RunEditor(args.firstFilename, args.secondFilename, true)
		if err != nil {
			side_channel.Panic(err)
		}
	}
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

func fileNonEmpty(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	if !info.Mode().IsRegular() {
		return false
	}
	return info.Size() > 0
}

func promptDeleteLogFile(args programArgs) bool {
	if fileNonEmpty(args.secondFilename) {
		ok := promptYesNo(fmt.Sprintf("log_writer file exists (%s), delete it?", args.secondFilename), false)
		if !ok {
			return false
		}
		err := os.Remove(args.secondFilename)
		if err != nil {
			side_channel.Panic(err)
		}
		return true
	}
	return true
}

func getDefaultLogFilename(inputFilename string) (firstFilename string, secondFilename string) {
	if fileNonEmpty(inputFilename) {
		firstFilename, _ = filepath.Abs(inputFilename)
		secondFilename = filepath.Join(config.Load().LOG_DIR, firstFilename)
	} else {
		firstFilename = ""
		secondFilename = filepath.Join(config.Load().LOG_DIR, "empty_file")
	}
	err := os.MkdirAll(filepath.Dir(secondFilename), 0o700)
	if err != nil {
		side_channel.Panic(err)
		return "", ""
	}

	return firstFilename, secondFilename
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
		pargs.firstFilename, pargs.secondFilename = getDefaultLogFilename(pargs.firstFilename)
	} else {
		pargs.firstFilename, _ = getDefaultLogFilename(pargs.firstFilename)
	}
	return pargs
}
