package chrome

type set[T comparable] map[T]struct{}

func (s set[T]) Add(k T) {
	s[k] = struct{}{}
}

func (s set[T]) Del(k T) {
	delete(s, k)
}

func (s set[T]) Has(k T) bool {
	_, ok := s[k]
	return ok
}
