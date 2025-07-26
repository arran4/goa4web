package dlqdefaults

import (
	dlqpkg "github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/dlq/db"
	"github.com/arran4/goa4web/internal/dlq/dir"
	"github.com/arran4/goa4web/internal/dlq/email"
	"github.com/arran4/goa4web/internal/dlq/file"
	emailpkg "github.com/arran4/goa4web/internal/email"
)

// RegisterDefaults registers all stable DLQ providers.
func RegisterDefaults(r *dlqpkg.Registry, er *emailpkg.Registry) {
	file.Register(r)
	dir.Register(r)
	db.Register(r)
	email.Register(r, er)
	dlqpkg.RegisterLogDLQ(r)
}

// NewRegistry returns a Registry with stable providers registered.
func NewRegistry(er *emailpkg.Registry) *dlqpkg.Registry {
	r := dlqpkg.NewRegistry()
	RegisterDefaults(r, er)
	return r
}
