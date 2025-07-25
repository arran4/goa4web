package dbdefaults

import (
	dbdrivers "github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dbdrivers/mysql"
	"github.com/arran4/goa4web/internal/dbdrivers/postgres"
	"github.com/arran4/goa4web/internal/dbdrivers/sqlite"
)

// Register registers all stable database connectors.
func Register(r *dbdrivers.Registry) {
	mysql.Register(r)
	postgres.Register(r)
	sqlite.Register(r)
}
