package model

type Model interface {
	Get(i int) []rune
	Set(i int, val []rune) Model
	Ins(i int, val []rune) Model
	Del(i int) Model
	Iter(f func(i int, val []rune) bool)
	Len() int
	Append(line Line) Model
}
