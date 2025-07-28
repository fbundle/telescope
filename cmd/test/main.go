package main

import (
	"fmt"
	"telescope/persistent/vector"
)

func main() {
	v := vector.New[int]()
	v = v.Ins(v.Len(), 0)
	v = v.Ins(v.Len(), 1)
	v = v.Ins(v.Len(), 2)
	v = v.Ins(v.Len(), 3)

	for i, val := range v.Iter {
		fmt.Println(i, val)
		if i >= 2 {
			break
		}
	}
}
