package side_channel

import (
	"fmt"
	"os"
	"runtime"
	"sync"
)

var (
	writeMu    sync.Mutex = sync.Mutex{}
	writeCount bool       = false
)

func writeln(vs []any, msg string) bool {
	sideChannelPath, ok := os.LookupEnv("SIDE_CHANNEL_PATH")
	if !ok {
		sideChannelPath = ".side_channel.log"
	}
	writeMu.Lock()
	defer writeMu.Unlock()

	if !writeCount { // first call
		writeCount = true
		_ = os.Remove(sideChannelPath)
	}

	prepend := func(vs []any, v any) []any {
		return append([]any{v}, vs...)
	}
	f, err := os.OpenFile(sideChannelPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return false
	}
	defer f.Close()
	_, _ = fmt.Fprintln(f, prepend(vs, msg)...)
	return true
}

func Panic(vs ...any) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		writeln(vs, "exit from unknown location")
		os.Exit(1)
	}

	writeln(vs, fmt.Sprintf("%s:%d", file, line))
	os.Exit(1)
}

func WriteLn(vs ...any) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		writeln(vs, "exit from unknown location")
		os.Exit(1)
	}
	ok = writeln(vs, fmt.Sprintf("%s:%d", file, line))
	if !ok {
		os.Exit(1)
	}
}
