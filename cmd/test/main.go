package main

import (
	"fmt"
	"telescope/util/persistent/ordered_map"
)

func main() {
	m := ordered_map.EmptyOrderedMap[int, string]()
	m = m.Set(1, "one")
	m = m.Set(2, "two")
	m = m.Set(3, "three")
	m = m.Set(2, "hai")
	m = m.Del(3)
	fmt.Println(m.Repr())
}
