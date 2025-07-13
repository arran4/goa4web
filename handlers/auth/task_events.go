package auth

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

// RegisterTask represents user registration.
var RegisterTask = eventbus.TaskEvent{
	Name:    hcommon.TaskRegister,
	Matcher: hcommon.TaskMatcher(hcommon.TaskRegister),
}
