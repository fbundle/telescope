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

const VERSION = "0.1.4"

func printHelp() {
	printVersion()
	help := `
Usage: "telescope [option] <file>"
Option:
  -h --help	: show help
  -v --version	: get version
  -r --replay	: replay the edited file 
	`
	fmt.Println(help)
}

func printVersion() {
	fmt.Printf("telescope version %s\n", VERSION)
}

func consume(args []string) ([]string, string) {
	if len(args) == 0 {
		return []string{}, ""
	}
	return args[1:], args[0]
}

func main() {
	var replay bool
	var inputFilename, journalFilename string

	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}

	args, head := consume(args)
	if head == "-h" || head == "--help" {
		printHelp()
		return
	}
	if head == "-v" || head == "--version" {
		printVersion()
		return
	}

	if head == "-r" || head == "--replay" {
		replay = true
		args, inputFilename = consume(args)
		args, journalFilename = consume(args)
	} else {
		inputFilename = head
		args, journalFilename = consume(args)
	}
	if len(journalFilename) == 0 {
		journalFilename = getDefaultJournalFilename(inputFilename)
	}

	if replay {
		err := app.RunReplay(inputFilename, journalFilename)
		if err != nil {
			panic(err)
		}
	} else {
		if fileExists(journalFilename) && fileSize(journalFilename) > 0 {
			ok := promptYesNo(fmt.Sprintf("journal file exists (%s), delete it?", journalFilename), false)
			if !ok {
				return
			}
			err := os.Remove(journalFilename)
			if err != nil {
				panic(err)
			}
		}
		err := app.RunEditor(inputFilename, journalFilename)
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
func getDefaultJournalFilename(filenameTextIn string) string {
	dir := filepath.Dir(filenameTextIn)
	name := "." + filepath.Base(filenameTextIn) + ".journal"
	journalPath := filepath.Join(dir, name)
	return journalPath
}
