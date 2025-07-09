//go:build !sqlite

package sqlite

import "log"

// Register logs that SQLite is disabled for quick tests.
func Register() {
	log.Println("sqlite: driver disabled for quick tests")
}
