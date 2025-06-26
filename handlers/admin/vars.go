package admin

import (
	"database/sql"
	"github.com/arran4/goa4web/pkg/server"
)

// ConfigFile points to the application's config file.
var ConfigFile string

// Srv holds the running server instance.
var Srv *server.Server

// DBPool exposes the database connection pool.
var DBPool *sql.DB
