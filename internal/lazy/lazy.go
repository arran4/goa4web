package lazy

import (
	"fmt"
	"sync"
)

// Value manages on-demand loading of a value.
type Value[T any] struct {
	once   sync.Once
	value  T
	err    error
	loaded bool
}

// load executes fn only once and caches its result.
func (l *Value[T]) Load(fn func() (T, error)) (T, error) {
	l.once.Do(func() {
		l.value, l.err = fn()
		l.loaded = true
	})
	return l.value, l.err
}

// set stores a precomputed value if not already loaded.
func (l *Value[T]) Set(v T) {
	l.once.Do(func() {
		l.value = v
		l.loaded = true
	})
}

// peek returns the cached value and whether it has been loaded.
func (l *Value[T]) Peek() (T, bool) {
	return l.value, l.loaded
}

// args holds the behaviour modifiers for lazy operations.
type args[T any] struct {
	dontFetch    bool
	refresh      bool
	clear        bool
	must         bool
	mustCached   bool
	setID        *int32
	setValue     *T
	defaultValue *T
}

// Option configures lazy retrieval behaviour.
type Option[T any] func(*args[T])

// DontFetch prevents data fetching when not already cached.
func DontFetch[T any]() Option[T] { return func(a *args[T]) { a.dontFetch = true } }

// Set stores v in the cache for the referenced ID.
func Set[T any](v T) Option[T] { return func(a *args[T]) { a.setValue = &v } }

// SetID changes which ID is referenced by the operation.
func SetID[T any](id int32) Option[T] { return func(a *args[T]) { a.setID = &id } }

// Refresh forces reloading the value by discarding the cached data.
func Refresh[T any]() Option[T] { return func(a *args[T]) { a.refresh = true } }

// Clear removes any cached value for the referenced ID.
func Clear[T any]() Option[T] { return func(a *args[T]) { a.clear = true } }

// MustBeCached errors if the value was not already cached.
func MustBeCached[T any]() Option[T] { return func(a *args[T]) { a.mustCached = true } }

// Must errors if fetching fails.
func Must[T any]() Option[T] { return func(a *args[T]) { a.must = true } }

// DefaultValue returns v when the lookup would otherwise yield nothing.
func DefaultValue[T any](v T) Option[T] { return func(a *args[T]) { a.defaultValue = &v } }

// Map performs a lazy lookup of id using fetch, caching the result in m.
func Map[T any](m *map[int32]*Value[T], id int32, fetch func(int32) (T, error), opts ...Option[T]) (T, error) {
	var zero T
	args := &args[T]{}
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
		*m = make(map[int32]*Value[T])
	}
	if args.clear {
		delete(*m, id)
		return zero, nil
	}
	lv, ok := (*m)[id]
	if !ok || args.refresh {
		lv = &Value[T]{}
		(*m)[id] = lv
	}
	if args.setValue != nil {
		lv.Set(*args.setValue)
		return *args.setValue, nil
	}
	v, loaded := lv.Peek()
	if loaded {
		return v, nil
	}
	if args.dontFetch {
		if args.mustCached && !loaded {
			return zero, fmt.Errorf("value not cached")
		}
		if args.defaultValue != nil {
			lv.Set(*args.defaultValue)
			return *args.defaultValue, nil
		}
		return v, nil
	}
	if fetch == nil {
		return zero, nil
	}
	v, err := lv.Load(func() (T, error) { return fetch(id) })
	if err != nil {
		if args.defaultValue != nil && !args.must {
			lv.Set(*args.defaultValue)
			return *args.defaultValue, nil
		}
		if args.must {
			return v, fmt.Errorf("fetch error: %w", err)
		}
		return v, err
	}
	return v, nil
}
