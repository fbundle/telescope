package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	jsonData := "{\"name\": \"Alice\", \"age\": 30}\n{\"name\": \"khanh\", \"age\": 31}aaaa"
	reader := strings.NewReader(jsonData)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	var person Person
	for {
		if err := decoder.Decode(&person); err != nil {
			buffer, _ := io.ReadAll(decoder.Buffered())
			rest, _ := io.ReadAll(reader)
			fmt.Println("the rest", string(buffer), string(rest), err)
			panic(err)
		}
		fmt.Println(person.Name, person.Age)
	}
}
