//go:build !sqlite

package sqlite

import (
	dbdrivers "github.com/arran4/goa4web/internal/dbdrivers"
	"log"
)

// Register logs that SQLite is disabled for quick tests.
func Register(r *dbdrivers.Registry) {
	log.Println("sqlite: driver disabled for quick tests")
}
