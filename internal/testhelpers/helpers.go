package testhelpers

import (
	"fmt"
)

// Must ensures that a function returning (T, error) did not return an error.
// If it did, it panics.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("Must failed: %v", err))
	}
	return v
}
