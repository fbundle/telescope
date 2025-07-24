package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"telescope/app"
	"telescope/feature"
	"telescope/journal"
)

const VERSION = "0.1.3"

func printHelp() {
	printVersion()
	help := `
Usage: telescope [option] <input_file> <output_file>
Option:
  -h --help	: show help
  -v --version	: get version
  -r --replay	: replay journal file
	`
	fmt.Println(help)
}

func printVersion() {
	fmt.Printf("telescope_extra version %s\n", VERSION)
}

func consume(args []string) ([]string, string) {
	if len(args) == 0 {
		return []string{}, ""
	}
	return args[1:], args[0]
}

func main() {
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

	var inputFilename, outputFilename, journalFilename string

	isRecover := head == "-r" || head == "--replay"
	if isRecover {
		args, inputFilename = consume(args)
		args, outputFilename = consume(args)
	} else {
		inputFilename = head
		args, outputFilename = consume(args)
	}

	// recover
	if isRecover {
		journalFilename = journal.GetJournalFilename(inputFilename)
		err := app.RunReplay(inputFilename, journalFilename)
		if err != nil {
			panic(err)
		}
		return
	}

	// text editor
	if feature.DisableJournal() {
		journalFilename = ""
	} else {
		journalFilename = journal.GetJournalFilename(inputFilename)
	}

	if len(journalFilename) > 0 {
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
	}

	err := app.RunEditor(inputFilename, journalFilename, outputFilename)
	if err != nil {
		log.Fatal(err)
	}
}

func recoverFromJournal() {

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
