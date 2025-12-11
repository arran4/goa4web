package database

import (
	_ "embed"
)

//go:embed schema.mysql.sql
var SchemaMySQL []byte

//go:embed seed.sql
var SeedSQL []byte
