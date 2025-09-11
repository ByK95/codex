package iterator

// Iterator holds state for generic iteration.
type Iterator[T any] struct {
	data []T
	curr int
}

// NewIterator creates an iterator over a slice of T.
func NewIterator[T any](items []T) *Iterator[T] {
	return &Iterator[T]{
		data: items,
		curr: 0,
	}
}

// Next returns the next element and true if available,
// or the zero value and false if iteration ended.
func (it *Iterator[T]) Next() (T, bool) {
	if it.curr >= len(it.data) {
		var zero T
		return zero, false
	}
	item := it.data[it.curr]
	it.curr++
	return item, true
}

// Reset allows reusing the iterator.
func (it *Iterator[T]) Reset() {
	it.curr = 0
}
