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
    cfg.SessionName = "goa4web_session"

	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	core.Store = store
	core.SessionName = cfg.SessionName

	s, err := app.NewServer(context.Background(), cfg, nil, app.WithDB(&db.QuerierStub{}))
    if err != nil {
        t.Fatalf("Failed to start server: %v", err)
    }

    r := s.Router

	// A purely anonymous route that doesn't need CSRF state.
	// Since we can't inject a route into the existing MUX cleanly without breaking its 404,
    // actually we can cast s.Router to *mux.Router!

    // BUT WAIT: s.Router is `http.Handler`. Wait, it's a `mux.Router` internally!
    // But let's check `internal/app/server/server.go`. It says `Router http.Handler`.
    // Actually, `app.NewServer` wraps it in middlewares! So `s.Router` is NOT a mux.Router.
    // The `server.Server` struct from `app.NewServer` has the full stack.
}
INNER_EOF
