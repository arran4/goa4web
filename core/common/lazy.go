package common

import "sync"

// lazyValue manages on-demand loading of a value.
type lazyValue[T any] struct {
	once   sync.Once
	value  T
	err    error
	loaded bool
}

// load executes fn only once and caches its result.
func (l *lazyValue[T]) load(fn func() (T, error)) (T, error) {
	l.once.Do(func() {
		l.value, l.err = fn()
		l.loaded = true
	})
	return l.value, l.err
}

// set stores a precomputed value if not already loaded.
func (l *lazyValue[T]) set(v T) {
	l.once.Do(func() {
		l.value = v
		l.loaded = true
	})
}

// peek returns the cached value and whether it has been loaded.
func (l *lazyValue[T]) peek() (T, bool) {
	return l.value, l.loaded
}
