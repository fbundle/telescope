package seq

const (
	delta = 3
)

// implement persistent sequence using weight-balanced tree

type node[T any] struct {
	weight uint64
	// height uint64
	entry T
	left  *node[T]
	right *node[T]
}

func height[T any](n *node[T]) uint64 {
	if n == nil {
		return 0
	}
	// return n.height
	return 0
}

func weight[T any](n *node[T]) uint64 {
	if n == nil {
		return 0
	}
	return n.weight
}

func makeNode[T any](entry T, left *node[T], right *node[T]) *node[T] {
	return &node[T]{
		weight: 1 + weight(left) + weight(right),
		// height: 1 + max(height(left), height(right)),
		entry: entry,
		left:  left,
		right: right,
	}
}

func get[T any](n *node[T], i uint64) T {
	if n == nil {
		panic("index out of range")
	}
	if i < weight(n.left) {
		return get(n.left, i)
	}
	if i < weight(n.left)+1 {
		return n.entry
	}
	if i < weight(n.left)+1+weight(n.right) {
		return get(n.right, i-(weight(n.left)+1))
	}
	panic("index out of range")
}

func iter[T any](n *node[T], f func(e T) bool) bool {
	if n == nil {
		return true // continue
	}

	ok := iter(n.left, f)
	if !ok {
		return false
	}
	ok = f(n.entry)
	if !ok {
		return false
	}
	return iter(n.right, f)
}

func balance[T any](n *node[T]) *node[T] {
	if n == nil {
		return nil
	}
	if weight(n.left)+weight(n.right) <= 1 {
		return n
	}
	if weight(n.left) > delta*weight(n.right) { // left is guaranteed to be non-nil
		// right rotate
		//         n
		//   l           r
		// ll lr
		//
		//      becomes
		//
		//         l
		//   ll          n
		//             lr r

		l, r := n.left, n.right
		ll, lr := l.left, l.right
		n1 := makeNode(n.entry, lr, r)
		l1 := makeNode(l.entry, ll, n1)
		return l1
	} else if delta*weight(n.left) < weight(n.right) { // right is guaranteed to be non-nil
		// left rotate
		//         n
		//   l           r
		//             rl rr
		//
		//      becomes
		//
		//         r
		//   n          rr
		//  l rl

		l, r := n.left, n.right
		rl, rr := r.left, r.right
		n1 := makeNode(n.entry, l, rl)
		r1 := makeNode(r.entry, n1, rr)
		return r1
	}
	return n
}

func set[T any](n *node[T], i uint64, entry T) *node[T] {
	if n == nil {
		panic("index out of range")
	}
	if i < weight(n.left) {
		l1 := set(n.left, i, entry)
		n1 := makeNode(n.entry, l1, n.right)
		return balance(n1)
	}
	if i < weight(n.left)+1 {
		n1 := makeNode(entry, n.left, n.right)
		return n1
	}
	if i < weight(n.left)+1+weight(n.right) {
		r1 := set(n.right, i-(weight(n.left)+1), entry)
		n1 := makeNode(n.entry, n.left, r1)
		return balance(n1)
	}
	panic("index out of range")
}

func ins[T any](n *node[T], i uint64, entry T) *node[T] {
	if n == nil && i > 0 {
		panic("index out of range")
	}
	if n == nil && i == 0 {
		return makeNode(entry, nil, nil)
	}
	if i < weight(n.left) {
		l1 := ins(n.left, i, entry)
		n1 := makeNode(n.entry, l1, n.right)
		return balance(n1)
	}
	if i < weight(n.left)+1 {
		r1 := ins(n.right, 0, n.entry)
		n1 := makeNode(entry, n.left, r1)
		return balance(n1)
	}
	if i <= weight(n.left)+1+weight(n.right) {
		r1 := ins(n.right, i-(weight(n.left)+1), entry)
		n1 := makeNode(n.entry, n.left, r1)
		return balance(n1)
	}
	panic("index out of range")
}

func del[T any](n *node[T], i uint64) *node[T] {
	if n == nil {
		panic("index out of range")
	}
	if i < weight(n.left) {
		l1 := del(n.left, i)
		n1 := makeNode(n.entry, l1, n.right)
		return balance(n1)
	}
	if i < weight(n.left)+1 {
		if weight(n.right) == 0 {
			return n.left
		}
		entry := get(n.right, 0)
		r1 := del(n.right, 0)
		n1 := makeNode(entry, n.left, r1)
		return balance(n1)
	}
	if i < weight(n.left)+1+weight(n.right) {
		r1 := del(n.right, i-(weight(n.left)+1))
		n1 := makeNode(n.entry, n.left, r1)
		return balance(n1)
	}
	panic("index out of range")
}

func merge[T any](l *node[T], r *node[T]) *node[T] {
	if l == nil {
		return r
	}
	if r == nil {
		return l
	}
	wl, wr := weight(l), weight(r)
	if wl > wr {
		entry := get(l, wl-1)
		l1 := del(l, wl-1)
		n1 := makeNode(entry, l1, r)
		return balance(n1)
	} else {
		entry := get(r, 0)
		r1 := del(r, 0)
		n1 := makeNode(entry, l, r1)
		return balance(n1)
	}
}

// split - ([0, 1, 2, 3, 4], 2) -> [0, 1] , [2, 3, 4]
func split[T any](n *node[T], i uint64) (*node[T], *node[T]) {
	if n == nil {
		return nil, nil
	}
	if i <= 0 {
		return nil, n
	}
	if i < weight(n.left) {
		ll1, lr1 := split(n.left, i)
		n1 := makeNode(n.entry, lr1, n.right)
		n2 := balance(n1)
		return ll1, n2
	}
	if i < weight(n.left)+1 {
		r1 := ins(n.right, 0, n.entry)
		return n.left, r1
	}
	if i < weight(n.left)+1+weight(n.right) {
		rl1, rr1 := split(n.right, i-(weight(n.left)+1))
		n1 := makeNode(n.entry, n.left, rl1)
		n2 := balance(n1)
		return n2, rr1
	}
	return n, nil
}
