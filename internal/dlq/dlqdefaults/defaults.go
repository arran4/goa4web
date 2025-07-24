package dlqdefaults

import (
	dlqpkg "github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/dlq/db"
	"github.com/arran4/goa4web/internal/dlq/dir"
	"github.com/arran4/goa4web/internal/dlq/email"
	"github.com/arran4/goa4web/internal/dlq/file"
	emailpkg "github.com/arran4/goa4web/internal/email"
)

// Register registers all stable DLQ providers.
func Register(d *dlqpkg.Registry, e *emailpkg.Registry) {
	file.Register(d)
	dir.Register(d)
	db.Register(d)
	email.Register(d, e)
}
