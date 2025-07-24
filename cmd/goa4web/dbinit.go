package main

import (
	"github.com/arran4/goa4web/internal/dbdrivers"
	dbdefaults "github.com/arran4/goa4web/internal/dbdrivers/dbdefaults"
)

func init() { dbdefaults.Register(dbdrivers.NewRegistry()) }
