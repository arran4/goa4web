//go:build sqlite || sqlite3

package dbdefaults

import (
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dbdrivers/sqlite"
)

func registerOptional(r *dbdrivers.Registry) {
	sqlite.Register(r)
}
