cat << 'INNER_EOF' > tests/csrf/csrf_prod_test.go
package csrf_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/middleware/csrf"
	"github.com/arran4/goa4web/internal/tasks"
    "github.com/arran4/goa4web/internal/app"
    "github.com/arran4/goa4web/internal/db"
    "github.com/arran4/goa4web/handlers"
)

func TestCSRFAnonymousPublicCacheable(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.SessionSecret = "testsecret"
    cfg.CSRFEnabled = true

	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	core.Store = store
	core.SessionName = cfg.SessionName

    // We can't easily spin up full DB here, but if app.NewServer requires dbConn,
    // maybe we can just build the middleware stack using app.NewServer.
    // Wait, the PR comment says "Please add a production-stack regression using the existing internal/app server/test harness (the SQLite end-to-end harness is the obvious place if that is the repository's normal path) with CSRF enabled."
}
INNER_EOF
