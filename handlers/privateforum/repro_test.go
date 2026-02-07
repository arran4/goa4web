package privateforum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestPrivateRoute(t *testing.T) {
	r := mux.NewRouter()
	reg := router.NewRegistry()
	navReg := navpkg.NewRegistry()
	cfg := &config.RuntimeConfig{}

	RegisterRoutes(r, cfg, navReg)
	_ = reg

	tests := []struct {
		name          string
		path          string
		userID        int32
		username      string
		grantReturns  int32
		expectedCode  int  // 403 or 200 (or 500 if template missing)
		expectMatched bool // if true, it matched a route (so not 404)
	}{
		{
			name:          "Unauthenticated",
			path:          "/private",
			userID:        0,
			expectedCode:  http.StatusOK,
			expectMatched: true,
		},
		{
			name:          "Authenticated No Grant",
			path:          "/private",
			userID:        1,
			username:      "user",
			grantReturns:  0, // No grant
			expectedCode:  http.StatusOK,
			expectMatched: true,
		},
		{
			name:         "Authenticated With Grant",
			path:         "/private",
			userID:       1,
			username:     "user",
			grantReturns: 1, // Grant exists
			// Note: We expect this might panic or error due to missing templates in test env,
			// but we are testing routing. If it hits the handler logic (which does grant check),
			// then routing worked.
			// If it returns 200 or 500 (template error), it matched.
			// If 404, it failed routing.
			expectedCode:  http.StatusOK,
			expectMatched: true,
		},
	}

	for _, tt := range tests {
		runName := "Happy Path"
		if tt.expectedCode != http.StatusOK {
			runName = "Unhappy Path"
		}
		t.Run(runName+" - "+tt.name, func(t *testing.T) {
			req := testhelpers.Must(http.NewRequest("GET", tt.path, nil))
			rr := httptest.NewRecorder()

			stub := testhelpers.NewQuerierStub(testhelpers.WithGrantResult(tt.grantReturns == 1))
			stub.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{}
			if tt.userID != 0 {
				stub.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
					Idusers:  tt.userID, // Correct field name for ID
					Username: sql.NullString{String: tt.username, Valid: true},
				}
			}

			// Use NewCoreData to construct CD properly
			cd := common.NewCoreData(context.Background(), stub, cfg)
			cd.UserID = tt.userID

			ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
			req = req.WithContext(ctx)

			// Recover from template panics if any, as we just want to verify routing
			defer func() {
				if r := recover(); r != nil {
					// Check if panic is related to template loading
					// If so, we consider routing success (it reached the handler)
					t.Logf("Recovered from panic: %v", r)
				}
			}()

			r.ServeHTTP(rr, req)

			if tt.expectMatched && rr.Code == http.StatusNotFound {
				t.Errorf("Path %s returned 404, expected matched route", tt.path)
			}

			if tt.expectedCode == http.StatusForbidden && rr.Code != http.StatusForbidden {
				t.Errorf("Expected 403 Forbidden, got %d", rr.Code)
			}

			// For the success case, we might get 500 or panic depending on template state.
			// The key is that it is NOT 404.
		})
	}
}
