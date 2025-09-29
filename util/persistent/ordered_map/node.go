package ordered_map

const (
	delta = 3
)

type Comparable[T any] interface {
	Cmp(T) int
}

// implement persistent ordered map using weight balanced tree

type node[T Comparable[T]] struct {
	weight uint64
	// height uint64
	entry T
	left  *node[T]
	right *node[T]
}

func height[T Comparable[T]](n *node[T]) uint64 {
	if n == nil {
		return 0
	}
	// return n.height
	return 0
}

func weight[T Comparable[T]](n *node[T]) uint64 {
	if n == nil {
		return 0
	}
	return n.weight
}
func makeNode[T Comparable[T]](entry T, left *node[T], right *node[T]) *node[T] {
	return &node[T]{
		weight: 1 + weight(left) + weight(right),
		// height: 1 + max(height(left), height(right)),
		entry: entry,
		left:  left,
		right: right,
	}
}

func get[T Comparable[T]](n *node[T], entryIn T) (entryOut T, ok bool) {
	if n == nil {
		return entryOut, false
	}
	cmp := n.entry.Cmp(entryIn)
	switch {
	case cmp < 0:
		return get(n.right, entryIn)
	case cmp > 0:
		return get(n.left, entryIn)
	default:
		return n.entry, true
	}
}

func iter[T Comparable[T]](n *node[T], f func(entryOut T) bool) bool {
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

func balance[T Comparable[T]](n *node[T]) *node[T] {
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

func set[T Comparable[T]](n *node[T], entryIn T) *node[T] {
	if n == nil {
		return makeNode(entryIn, nil, nil)
	}
	cmp := n.entry.Cmp(entryIn)
	switch {
	case cmp < 0: // n.entry < entryIn
		r1 := set(n.right, entryIn)
		n1 := makeNode(n.entry, n.left, r1)
		return balance(n1)
	case cmp > 0: // n.entry > entryIn
		l1 := set(n.left, entryIn)
		n1 := makeNode(n.entry, l1, n.right)
		return balance(n1)
	default: // n.entry == entryIn
		return makeNode(entryIn, n.left, n.right)
	}
}

func del[T Comparable[T]](n *node[T], entryIn T) *node[T] {
	if n == nil {
		return nil
	}
	cmp := n.entry.Cmp(entryIn)
	switch {
	case cmp < 0: // n.entry < entryIn
		r1 := del(n.right, entryIn)
		n1 := makeNode(n.entry, n.left, r1)
		return balance(n1)
	case cmp > 0: // n.entry > entryIn
		l1 := del(n.left, entryIn)
		n1 := makeNode(n.entry, l1, n.right)
		return balance(n1)
	default: // n.entry == entryIn
		return merge(n.left, n.right)
	}
}

func getMinEntry[T Comparable[T]](n *node[T]) T {
	if n == nil {
		panic("min of nil tree")
	}
	if n.left == nil {
		return n.entry
	}
	return getMinEntry(n.left)
}

func getMaxEntry[T Comparable[T]](n *node[T]) T {
	if n == nil {
		panic("max of nil tree")
	}
	if n.right == nil {
		return n.entry
	}
	return getMaxEntry(n.right)

}

func merge[T Comparable[T]](l *node[T], r *node[T]) *node[T] {
	if l == nil {
		return r
	}
	if r == nil {
		return l
	}
	wl, wr := weight(l), weight(r)
	if wl > wr {
		entry := getMaxEntry(l)
		l1 := del(l, entry)
		n1 := makeNode(entry, l1, r)
		return balance(n1)
	} else {
		entry := getMinEntry(r)
		r1 := del(r, entry)
		n1 := makeNode(entry, l, r1)
		return balance(n1)
	}
}

// split - ([1, 2, 3, 4], 3) -> [1, 2] , [3, 4]
func split[T Comparable[T]](n *node[T], entryIn T) (*node[T], *node[T]) {
	if n == nil {
		return nil, nil
	}
	cmp := n.entry.Cmp(entryIn)
	switch {
	case cmp < 0: // n.entry < entryIn
		rl1, rr1 := split(n.right, entryIn)
		n1 := makeNode(n.entry, n.left, rl1)
		n2 := balance(n1)
		return n2, rr1
	case cmp > 0: // n.entry > entryIn
		ll1, lr1 := split(n.left, entryIn)
		n1 := makeNode(n.entry, lr1, n.right)
		n2 := balance(n1)
		return ll1, n2
	default: // n.entry == entryIn
		return n.left, set(n.right, n.entry)
	}
}
