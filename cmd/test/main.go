package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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
func main() {
	var ok bool
	ok = promptYesNo("ues or no", false)
	fmt.Println(ok)
	ok = promptYesNo("ues or no", true)
	fmt.Println(ok)
}
