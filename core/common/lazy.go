package common

import (
	"fmt"
	"sync"
)

// TODO extract lazy system into a reusable library.

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

// lazyArgs holds the behaviour modifiers for lazy operations.
type lazyArgs[T any] struct {
	dontFetch    bool
	refresh      bool
	clear        bool
	must         bool
	mustCached   bool
	setID        *int32
	setValue     *T
	defaultValue *T
}

// LazyOption configures lazy retrieval behaviour.
type LazyOption[T any] func(*lazyArgs[T])

// LazyDontFetch prevents data fetching when not already cached.
func LazyDontFetch[T any]() LazyOption[T] { return func(a *lazyArgs[T]) { a.dontFetch = true } }

// LazySet stores v in the cache for the referenced ID.
func LazySet[T any](v T) LazyOption[T] { return func(a *lazyArgs[T]) { a.setValue = &v } }

// LazySetID changes which ID is referenced by the operation.
func LazySetID[T any](id int32) LazyOption[T] { return func(a *lazyArgs[T]) { a.setID = &id } }

// LazyRefresh forces reloading the value by discarding the cached data.
func LazyRefresh[T any]() LazyOption[T] { return func(a *lazyArgs[T]) { a.refresh = true } }

// LazyClear removes any cached value for the referenced ID.
func LazyClear[T any]() LazyOption[T] { return func(a *lazyArgs[T]) { a.clear = true } }

// LazyMustBeCached errors if the value was not already cached.
func LazyMustBeCached[T any]() LazyOption[T] { return func(a *lazyArgs[T]) { a.mustCached = true } }

// LazyMust errors if fetching fails.
func LazyMust[T any]() LazyOption[T] { return func(a *lazyArgs[T]) { a.must = true } }

// LazyDefaultValue returns v when the lookup would otherwise yield nothing.
func LazyDefaultValue[T any](v T) LazyOption[T] { return func(a *lazyArgs[T]) { a.defaultValue = &v } }

func lazyMap[T any](m *map[int32]*lazyValue[T], id int32, fetch func(int32) (T, error), opts ...LazyOption[T]) (T, error) {
	var zero T
	args := &lazyArgs[T]{}
	for _, opt := range opts {
		opt(args)
	}
	if args.setID != nil {
		id = *args.setID
	}
	if m == nil {
		return zero, fmt.Errorf("lazy map pointer nil")
	}
	if *m == nil {
		*m = make(map[int32]*lazyValue[T])
	}
	if args.clear {
		delete(*m, id)
		return zero, nil
	}
	lv, ok := (*m)[id]
	if !ok || args.refresh {
		lv = &lazyValue[T]{}
		(*m)[id] = lv
	}
	if args.setValue != nil {
		lv.set(*args.setValue)
		return *args.setValue, nil
	}
	v, loaded := lv.peek()
	if loaded {
		return v, nil
	}
	if args.dontFetch {
		if args.mustCached && !loaded {
			return zero, fmt.Errorf("value not cached")
		}
		if args.defaultValue != nil {
			lv.set(*args.defaultValue)
			return *args.defaultValue, nil
		}
		return v, nil
	}
	if fetch == nil {
		return zero, nil
	}
	v, err := lv.load(func() (T, error) { return fetch(id) })
	if err != nil {
		if args.defaultValue != nil && !args.must {
			lv.set(*args.defaultValue)
			return *args.defaultValue, nil
		}
		if args.must {
			return v, fmt.Errorf("fetch error: %w", err)
		}
		return v, err
	}
	return v, nil
}
