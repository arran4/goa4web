package emaildefaults

import (
	emailpkg "github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/email/jmap"
	"github.com/arran4/goa4web/internal/email/local"
	"github.com/arran4/goa4web/internal/email/log"
	"github.com/arran4/goa4web/internal/email/ses"
	"github.com/arran4/goa4web/internal/email/smtp"
)

// Register registers all stable email providers.
func Register(r *emailpkg.Registry) {
	smtp.Register(r)
	ses.Register(r)
	jmap.Register(r)
	local.Register(r)
	log.Register(r)
}
