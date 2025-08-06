package main

import (
	"errors"
	"fmt"
	"telescope/util/monad"
)

func main() {
	m := monad.Unit[int](2, 3, 4)
	m = monad.Bind(m, func(t1 int) (*monad.Monad[int], error) {
		if t1 == 4 {
			return nil, errors.New("my_error")
		}
		ts := make([]int, 0)
		for i := 0; i < t1; i++ {
			ts = append(ts, t1)
		}
		return monad.Unit[int](ts...), nil
	})
	fmt.Println(m.Unwrap())
}
