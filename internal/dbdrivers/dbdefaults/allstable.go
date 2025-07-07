package dbdefaults

import (
	_ "github.com/arran4/goa4web/internal/dbdrivers/mysql"
	_ "github.com/arran4/goa4web/internal/dbdrivers/postgres"
	_ "github.com/arran4/goa4web/internal/dbdrivers/sqlite"
)

// Register registers all stable database connectors.
func Register() {}
