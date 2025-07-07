package dbdefaults

import (
	"github.com/arran4/goa4web/cmd/goa4web/dbhandlers/mysql"
	"github.com/arran4/goa4web/cmd/goa4web/dbhandlers/postgres"
	"github.com/arran4/goa4web/cmd/goa4web/dbhandlers/sqlite"
)

func init() {
	mysql.Register()
	postgres.Register()
	sqlite.Register()
}
