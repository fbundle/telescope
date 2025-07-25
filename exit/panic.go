package exit

import (
	"fmt"
	"os"
	"runtime"
)

func Write(vs ...any) {
	var buffer []byte

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		b := []byte(fmt.Sprintf("exit from %s:%d\n", file, line))
		buffer = append(buffer, b...)
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			b := []byte(fmt.Sprintf("function name: %s\n", fn.Name()))
			buffer = append(buffer, b...)
		}
	}

	b := []byte(fmt.Sprint(vs...))
	buffer = append(buffer, b...)

	os.WriteFile("exit.txt", buffer, 0600)
	os.Exit(1)
}
