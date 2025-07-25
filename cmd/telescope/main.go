package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"telescope/app"
)

const VERSION = "0.1.5"

func printHelp() {
	fmt.Printf("telescope version %s\n", VERSION)
	help := `
Usage: "telescope [option] file [logfile]"
Option:
  -h --help	: show help
  -v --version	: get version
  -r --replay	: replay the edited file 
  -l --log		: print the human readable log format
	`
	fmt.Println(help)
}

func printVersion() {
	fmt.Println(VERSION)
}

type programArgs struct {
	option        string
	inputFilename string
	logFilename   string
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
		if err := app.RunReplay(args.inputFilename, args.logFilename); err != nil {
			log.Fatalln(err)
		}
	case "-l", "--log":
		err := app.RunLog(args.logFilename)
		if err != nil {
			panic(err)
		}
	default:
		if fileExists(args.logFilename) && fileSize(args.logFilename) > 0 {
			ok := promptYesNo(fmt.Sprintf("log file exists (%s), delete it?", args.logFilename), false)
			if !ok {
				return
			}
			err := os.Remove(args.logFilename)
			if err != nil {
				panic(err)
			}
		}
		err := app.RunEditor(args.inputFilename, args.logFilename)
		if err != nil {
			panic(err)
		}
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func fileSize(filename string) int {
	info, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	size := info.Size() // in bytes
	return int(size)
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
func getDefaultLogFilename(filenameTextIn string) string {
	dir := filepath.Dir(filenameTextIn)
	name := "." + filepath.Base(filenameTextIn) + ".log"
	logPath := filepath.Join(dir, name)
	return logPath
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
	args, pargs.inputFilename = consume(args)
	args, pargs.logFilename = consume(args)
	if len(pargs.logFilename) == 0 {
		pargs.logFilename = getDefaultLogFilename(pargs.inputFilename)
	}
	return pargs
}
