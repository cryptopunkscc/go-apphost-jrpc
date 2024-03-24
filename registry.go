package jrpc

type Registry[V any] struct {
	next  map[byte]*Registry[V]
	value V
	empty V
}

func NewRegistry[V any]() *Registry[V] {
	return &Registry[V]{next: make(map[byte]*Registry[V])}
}

func (n *Registry[V]) Add(str string, v V) {
	if len(str) == 0 {
		n.value = v
		return
	}
	var next *Registry[V]
	next, ok := n.next[str[0]]
	if !ok {
		next = NewRegistry[V]()
		n.next[str[0]] = next
	}
	next.Add(str[1:], v)
}

func (n *Registry[V]) Unfold(str string) (string, V) {
	if len(str) == 0 {
		switch len(n.next) {
		case 0:
			return str, n.value
		case 1:
			for _, n := range n.next {
				return n.Unfold(str)
			}
		}
	}
	nn, ok := n.next[str[0]]
	if !ok {
		return str, n.value
	}
	return nn.Unfold(str[1:])
}

func (n *Registry[V]) All(str []byte, m map[string]V) {
	for b, r := range n.next {
		s := append(str, b)
		var value any = r.value
		var empty any = r.empty
		if value != empty {
			m[string(s)] = r.value
		}
		if len(r.next) > 0 {
			r.All(s, m)
		}
	}
}
