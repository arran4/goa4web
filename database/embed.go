package database

import (
	_ "embed"
	"strings"
)

//go:embed schema.mysql.sql
var SchemaMySQL []byte

//go:embed seed.sql
var SeedSQL []byte

//go:embed seed.sqlite.sql
var SeedSQLiteSQL []byte

// SeedSQLForDriver returns the seed SQL for the specified database driver.
func SeedSQLForDriver(driver string) []byte {
	switch strings.ToLower(driver) {
	case "sqlite", "sqlite3":
		if len(SeedSQLiteSQL) > 0 {
			return SeedSQLiteSQL
		}
	}
	return SeedSQL
}
