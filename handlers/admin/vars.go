package admin

import (
	"database/sql"
	"time"

	"github.com/arran4/goa4web/internal/app/server"

	"github.com/arran4/goa4web/core"
)

// ConfigFile points to the application's config file.
var ConfigFile string

// Srv holds the running server instance.
var Srv *server.Server

// DBPool exposes the database connection pool.
var DBPool *sql.DB

// AdminAPISecret is used to sign and verify administrator API tokens.
var AdminAPISecret string

// UpdateConfigKeyFunc is used to persist configuration changes. It should be
// set by the main application on startup.
var UpdateConfigKeyFunc func(fs core.FileSystem, path, key, value string) error

// StartTime marks when the server began running.
var StartTime time.Time
