package main

import (
	"context"
	"fmt"
	"telescope/journal"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	w, err := journal.NewWriter[journal.Entry](ctx, "test.json")
	if err != nil {
		panic(err)
	}
	w.Write(journal.Entry{
		Command:   "type",
		Rune:      'h',
		CursorRow: 1,
		CursorCol: 1,
	}).Write(journal.Entry{
		Command:   "type",
		Rune:      'e',
		CursorRow: 1,
		CursorCol: 1,
	})

	cancel()
	time.Sleep(30 * time.Second)

	err = journal.Read(context.Background(), "test.json", func(entry journal.Entry) {
		fmt.Println(entry)
	}, func() {
		fmt.Println("done")
	})
	if err != nil {
		panic(err)
	}

}
