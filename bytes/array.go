package bytes

type Array interface {
	Len() int
	At(i int) byte
	ReadAt(b []byte, i int64) (n int, err error)
}
