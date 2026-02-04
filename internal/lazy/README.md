# Lazy Evaluation Package

The `lazy` package provides generic, thread-safe primitives for lazy evaluation and caching of values. It is designed to handle expensive initialization operations that should only be performed once, or to manage caches of items loaded on demand.

## Features

- **Thread-Safe**: Uses `sync.Once` and atomic operations to ensure values are initialized exactly once, even under concurrent access.
- **Generics**: Fully supports Go generics for type safety.
- **Flexible Mapping**: Includes a helper for managing lazily loaded values in a map.
- **Configurable**: extensive options for controlling fetch behavior (timeouts, defaults, forced refreshes, etc.).

## Usage

### Single Value

The `Value[T]` struct allows you to lazily load a single value.

```go
package main

import (
	"fmt"
	"github.com/arran4/goa4web/internal/lazy"
)

func main() {
	var config lazy.Value[map[string]string]

	// The initialization function is only called once.
	val, err := config.Load(func() (map[string]string, error) {
		fmt.Println("Loading config...")
		return map[string]string{"key": "value"}, nil
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(val["key"]) // Output: value

	// Subsequent calls return the cached value immediately.
	cachedVal, _ := config.Load(nil)
	fmt.Println(cachedVal["key"]) // Output: value
}
```

### Lazy Map

The `Map` function provides a convenient way to manage a collection of lazy values, keyed by an integer ID (e.g., database IDs). It handles map locking and value initialization.

```go
package main

import (
	"fmt"
	"sync"
	"github.com/arran4/goa4web/internal/lazy"
)

func main() {
	// The map that holds the cache.
	cache := make(map[int32]*lazy.Value[string])
	var mu sync.Mutex

	fetchUser := func(id int32) (string, error) {
		fmt.Printf("Fetching user %d\n", id)
		return fmt.Sprintf("User-%d", id), nil
	}

	// Fetch user 1 (will trigger fetch)
	u1, err := lazy.Map(&cache, &mu, 1, fetchUser)
	if err != nil {
		panic(err)
	}
	fmt.Println(u1)

	// Fetch user 1 again (will use cache)
	u1Cached, _ := lazy.Map(&cache, &mu, 1, fetchUser)
	fmt.Println(u1Cached)

	// Options can modify behavior, e.g., force refresh
	u1Refreshed, _ := lazy.Map(&cache, &mu, 1, fetchUser, lazy.Refresh[string]())
	fmt.Println(u1Refreshed)
}
```

## API Overview

### Types

- `Value[T]`: The core struct for lazy loading. Zero value is ready to use.
- `Option[T]`: Functional options for `Map`.

### Functions

- `Map`: manages looking up, creating, and loading values in a map.

### Options for Map

- `DontFetch`: Returns the cached value if present, otherwise zero/default (does not trigger fetch).
- `Set`: Manually sets the value for the key.
- `SetID`: Overrides the ID used for lookup.
- `Refresh`: Forces a reload of the value.
- `Clear`: Removes the value from the map.
- `Must`: Wraps errors from the fetch function.
- `MustBeCached`: Returns an error if the value is not already cached.
- `DefaultValue`: Returns this value if lookup fails or (optionally) if fetch fails.

## Thread Safety

- **Value[T]**: `Load`, `Set`, and `Peek` are safe for concurrent use. `Load` guarantees the initialization function runs exactly once.
- **Map**: Requires the caller to provide a `sync.Mutex` which it uses to protect map operations (insertion/deletion). The value loading itself happens outside the map lock to avoid blocking other lookups, utilizing the internal safety of `Value[T]`.
