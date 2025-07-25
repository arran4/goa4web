package uploaddefaults

import (
	"github.com/arran4/goa4web/internal/upload"
	"github.com/arran4/goa4web/internal/upload/local"
	s3pkg "github.com/arran4/goa4web/internal/upload/s3"
)

// Register registers all built-in upload providers.
func Register(r *upload.Registry) {
	local.Register(r)
	if s3pkg.Built {
		s3pkg.Register(r)
	}
}
