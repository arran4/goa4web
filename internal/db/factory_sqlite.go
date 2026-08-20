//go:build sqlite || sqlite3

package db

func newSQLiteQuerier(dbtx DBTX) Querier {
	return NewSQLiteQuerier(dbtx)
}
