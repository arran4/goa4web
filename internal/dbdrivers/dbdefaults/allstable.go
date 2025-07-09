package dbdefaults

import (
	"github.com/arran4/goa4web/internal/dbdrivers/mysql"
	"github.com/arran4/goa4web/internal/dbdrivers/postgres"
	"github.com/arran4/goa4web/internal/dbdrivers/sqlite"
)

// Register registers all stable database connectors.
func Register() {
	mysql.Register()
	postgres.Register()
	sqlite.Register()
}
