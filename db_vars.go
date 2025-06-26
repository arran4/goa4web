package goa4web

import "database/sql"

var (
	dbPool         *sql.DB
	dbLogVerbosity int
)
