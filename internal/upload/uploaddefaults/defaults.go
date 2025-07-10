package uploaddefaults

import "github.com/arran4/goa4web/internal/upload/local"

// Register registers all built-in upload providers.
func Register() {
	local.Register()
}
