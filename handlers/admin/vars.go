package admin

import (
	"time"
)

// AdminAPISecret is used to sign and verify administrator API tokens.
var AdminAPISecret string

// StartTime marks when the server began running.
var StartTime time.Time
