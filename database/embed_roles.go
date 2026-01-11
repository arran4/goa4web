package database

import "embed"

// RolesFS contains embedded role SQL scripts located in this directory's roles/ subfolder.
//
//go:embed roles/*.sql
var RolesFS embed.FS
