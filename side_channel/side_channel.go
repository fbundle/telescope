package side_channel

import (
	"fmt"
	"os"
	"runtime"
)

func write(vs []any, msg string) bool {
	sideChannelPath := "side_channel.txt"
	prepend := func(vs []any, v any) []any {
		return append([]any{v}, vs...)
	}
	f, err := os.OpenFile(sideChannelPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return false
	}
	defer f.Close()
	_, _ = fmt.Fprintln(f, prepend(vs, msg))
	return true
}

func Panic(vs ...any) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		write(vs, "exit from unknown location")
		os.Exit(1)
	}

	write(vs, fmt.Sprintf("%s:%d", file, line))
	os.Exit(1)
}

func Write(vs ...any) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		write(vs, "exit from unknown location")
		os.Exit(1)
	}
	ok = write(vs, fmt.Sprintf("%s:%d", file, line))
	if !ok {
		os.Exit(1)
	}
}
