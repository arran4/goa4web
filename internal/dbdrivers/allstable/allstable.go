package allstable

import (
	"github.com/arran4/goa4web/internal/dbdrivers/mysql"
	"github.com/arran4/goa4web/internal/dbdrivers/postgres"
	"github.com/arran4/goa4web/internal/dbdrivers/sqlite3"
)

// Register registers all stable database connectors.
func Register() {
	mysql.Register()
	postgres.Register()
	sqlite3.Register()
}
