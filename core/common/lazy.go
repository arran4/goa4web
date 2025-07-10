package common

import "sync"

// lazyValue manages on-demand loading of a value.
type lazyValue[T any] struct {
	once  sync.Once
	value T
	err   error
}

// load executes fn only once and caches its result.
func (l *lazyValue[T]) load(fn func() (T, error)) (T, error) {
	l.once.Do(func() { l.value, l.err = fn() })
	return l.value, l.err
}

// set stores a precomputed value if not already loaded.
func (l *lazyValue[T]) set(v T) {
	l.once.Do(func() { l.value = v })
}
