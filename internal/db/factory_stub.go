//go:build !sqlite && !sqlite3

package db

func newSQLiteQuerier(dbtx DBTX) Querier {
	panic("sqlite driver requested but Goa4Web was compiled without SQLite support (build with -tags sqlite or -tags sqlite3)")
}
