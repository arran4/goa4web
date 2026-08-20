package db

// NewForDriver returns the appropriate Querier implementation for the given database driver.
func NewForDriver(dbtx DBTX, driver string) Querier {
	switch driver {
	case "sqlite", "sqlite3", "sqlite3-modernc":
		return newSQLiteQuerier(dbtx)
	default:
		return New(dbtx)
	}
}
