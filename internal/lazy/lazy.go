package lazy

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Value manages a value that is loaded on demand.
// It guarantees that the initialization function is called only once,
// even if accessed concurrently.
// It uses sync.Once for synchronization and atomic.Bool for lock-free status checks.
type Value[T any] struct {
	once   sync.Once
	value  T
	err    error
	loaded atomic.Bool
}

// Load ensures the value is loaded by executing fn if it hasn't been loaded yet.
// Subsequent calls return the cached value and error.
// Safe for concurrent use.
func (l *Value[T]) Load(fn func() (T, error)) (T, error) {
	l.once.Do(func() {
		l.value, l.err = fn()
		l.loaded.Store(true)
	})
	return l.value, l.err
}

// Set manually sets the value if it hasn't been loaded yet.
// If the value is already loaded (via Load or Set), this operation is a no-op.
// Safe for concurrent use.
func (l *Value[T]) Set(v T) {
	l.once.Do(func() {
		l.value = v
		l.loaded.Store(true)
	})
}

// Peek returns the cached value and true if it has been loaded.
// If not loaded, it returns the zero value of T and false.
// Safe for concurrent use.
func (l *Value[T]) Peek() (T, bool) {
	if l.loaded.Load() {
		return l.value, true
	}
	var zero T
	return zero, false
}

// args holds the configuration for Map operations.
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

// Option configures the behavior of the Map function.
type Option[T any] func(*args[T])

// DontFetch returns an Option that prevents fetching the value if it's not in the cache.
// If the value is missing, Map will return the zero value (or DefaultValue if set) and no error.
func DontFetch[T any]() Option[T] { return func(a *args[T]) { a.dontFetch = true } }

// Set returns an Option that manually sets the value for the given ID in the map.
// This bypasses the fetch function.
func Set[T any](v T) Option[T] { return func(a *args[T]) { a.setValue = &v } }

// SetID returns an Option that overrides the ID used for the map lookup.
func SetID[T any](id int32) Option[T] { return func(a *args[T]) { a.setID = &id } }

// Refresh returns an Option that forces a reload of the value, discarding any cached entry.
func Refresh[T any]() Option[T] { return func(a *args[T]) { a.refresh = true } }

// Clear returns an Option that removes the value associated with the ID from the map.
func Clear[T any]() Option[T] { return func(a *args[T]) { a.clear = true } }

// MustBeCached returns an Option that causes Map to return an error if the value is not already cached.
// Typically used with DontFetch.
func MustBeCached[T any]() Option[T] { return func(a *args[T]) { a.mustCached = true } }

// Must returns an Option that wraps any error returned by the fetch function.
func Must[T any]() Option[T] { return func(a *args[T]) { a.must = true } }

// DefaultValue returns an Option that specifies a fallback value to return if the value is not found
// (when DontFetch is used) or if fetching fails (unless Must is also used).
func DefaultValue[T any](v T) Option[T] { return func(a *args[T]) { a.defaultValue = &v } }

// Map retrieves or creates a lazy Value in the provided map.
// It handles locking the map using the provided mutex.
//
// Parameters:
//   - m: Pointer to the map caching the values.
//   - mu: Mutex protecting the map.
//   - id: The key to look up in the map.
//   - fetch: Function to generate the value if not found.
//   - opts: Optional modifiers.
//
// Returns the value and any error encountered.
func Map[T any](m *map[int32]*Value[T], mu *sync.Mutex, id int32, fetch func(int32) (T, error), opts ...Option[T]) (T, error) {
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
	if mu == nil {
		return zero, fmt.Errorf("lazy map mutex nil")
	}
	mu.Lock()
	if *m == nil {
		*m = make(map[int32]*Value[T])
	}
	if args.clear {
		delete(*m, id)
		mu.Unlock()
		return zero, nil
	}
	lv, ok := (*m)[id]
	if !ok || args.refresh {
		lv = &Value[T]{}
		(*m)[id] = lv
	}
	mu.Unlock()

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
