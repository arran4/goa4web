package main

import (
	dbdefaults "github.com/arran4/goa4web/internal/dbdrivers/dbdefaults"
)

func init() {
	dbdefaults.Register()
}
