package main

import (
	"github.com/arran4/goa4web/internal/dbdrivers/allstable"
)

func init() {
	allstable.Register()
}
