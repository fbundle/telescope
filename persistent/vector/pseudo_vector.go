package vector

func NewPseudoVector[T any]() Vector[T] {
	return &pseudoVector[T]{}
}

type pseudoVector[T any] struct {
	vec []T
}

func (p *pseudoVector[T]) Get(i int) T {
	return p.vec[i]
}

func (p *pseudoVector[T]) Set(i int, val T) Vector[T] {
	vec := make([]T, len(p.vec))
	copy(vec, p.vec)
	vec[i] = val
	return &pseudoVector[T]{vec: vec}
}

func (p *pseudoVector[T]) Ins(i int, val T) Vector[T] {
	vec := make([]T, len(p.vec)+1)
	copy(vec[:i], p.vec[:i])
	copy(vec[i+1:], p.vec[i:])
	vec[i] = val
	return &pseudoVector[T]{vec: vec}
}

func (p *pseudoVector[T]) Del(i int) Vector[T] {
	vec := make([]T, len(p.vec)-1)
	copy(vec[:i], p.vec[:i])
	copy(vec[i:], p.vec[i+1:])
	return &pseudoVector[T]{vec: vec}
}

func (p *pseudoVector[T]) Iter(f func(val T) bool) {
	for _, val := range p.vec {
		if !f(val) {
			break
		}
	}
}

func (p *pseudoVector[T]) Len() int {
	return int(len(p.vec))
}

func (p *pseudoVector[T]) Height() int {
	return 1
}

func (p *pseudoVector[T]) Split(i int) (Vector[T], Vector[T]) {
	n1 := make([]T, i)
	n2 := make([]T, len(p.vec)-int(i))
	copy(n1, p.vec[:i])
	copy(n2, p.vec[i:])
	return &pseudoVector[T]{vec: n1}, &pseudoVector[T]{vec: n2}
}

func (p *pseudoVector[T]) Concat(other Vector[T]) Vector[T] {
	n1, n2 := p.vec, other.(*pseudoVector[T]).vec
	n3 := make([]T, len(n1)+len(n2))
	copy(n3, n1)
	copy(n3[len(n1):], n2)
	return &pseudoVector[T]{vec: n3}
}

func (p *pseudoVector[T]) Repr() []T {
	return p.vec
}
