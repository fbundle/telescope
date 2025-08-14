package main

import (
	"fmt"
	"telescope/util/iterator"
)

func echo(ch iterator.Iterator[int]) {
	for i := range ch {
		fmt.Println(i)
	}
}

func main() {
	echo(func(f func(int) bool) {
		for i := 0; i < 10; i++ {
			f(i)
		}
	})
}
