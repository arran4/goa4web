package dlqdefaults

import (
	"github.com/arran4/goa4web/internal/dlq/db"
	"github.com/arran4/goa4web/internal/dlq/dir"
	"github.com/arran4/goa4web/internal/dlq/email"
	"github.com/arran4/goa4web/internal/dlq/file"
)

// Register registers all stable DLQ providers.
func Register() {
	file.Register()
	dir.Register()
	db.Register()
	email.Register()
}
