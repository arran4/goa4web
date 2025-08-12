package dbdefaults

import (
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dbdrivers/mysql"
)

// Register registers all stable database connectors.
func Register(r *dbdrivers.Registry) {
	mysql.Register(r)
}
