package emaildefaults

import (
	"github.com/arran4/goa4web/internal/email/jmap"
	"github.com/arran4/goa4web/internal/email/local"
	"github.com/arran4/goa4web/internal/email/log"
	"github.com/arran4/goa4web/internal/email/ses"
	"github.com/arran4/goa4web/internal/email/smtp"
)

// Register registers all stable email providers.
func Register() {
	smtp.Register()
	ses.Register()
	jmap.Register()
	local.Register()
	log.Register()
}
