package main

import (
	"fmt"
	"telescope/util/monad"
)

func main() {
	m := monad.Unit[int](2, 3, 4)
	for v := range m.Chan {
		fmt.Println(v)
	}
	if m.Error != nil {
		fmt.Println(m.Error)
	}
}
