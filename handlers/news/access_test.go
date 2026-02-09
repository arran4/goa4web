package news

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func TestNewsPostAccessControl(t *testing.T) {
	// Setup Session Store
	store := sessions.NewCookieStore([]byte("test-secret"))
	core.Store = store
	core.SessionName = "test-session"

	// Setup Router
	r := mux.NewRouter()
	navReg := navpkg.NewRegistry()
	RegisterRoutes(r, config.NewRuntimeConfig(), navReg)

	t.Run("Access Denied - No Grant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/news/news/123", nil)

		// Inject CoreData with NO grants and UserID=1 (logged in)
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
		cd.UserID = 1 // Logged in user

		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		// Expect 403 (RequireNewsPostView returns styled error page)
		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden (denied by wrapper), got %d", rr.Code)
		}
	})

	t.Run("Access Granted - With Grant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/news/news/123", nil)

		// Inject grant and UserID=1
		grants := []*db.Grant{
			{
				Section: "news",
				Item:    sql.NullString{String: "post", Valid: true},
				Action:  "view",
				ItemID:  sql.NullInt32{Int32: 123, Valid: true},
				Active:  true,
			},
		}
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithGrants(grants))
		cd.UserID = 1

		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		// Expect 403 (RequireNewsPostView passes, Handler runs, Handler fails due to DB missing data)
		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden (handler ran but missing data), got %d", rr.Code)
		}
	})
}
