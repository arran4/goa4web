package dlqdefaults

import (
	dlqpkg "github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/dlq/db"
	"github.com/arran4/goa4web/internal/dlq/dir"
	"github.com/arran4/goa4web/internal/dlq/email"
	"github.com/arran4/goa4web/internal/dlq/file"
)

// Register registers all stable DLQ providers.
func Register(r *dlqpkg.Registry) {
	file.Register(r)
	dir.Register(r)
	db.Register(r)
	email.Register(r)
}
