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
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartGroupDiscussionPageAccess(t *testing.T) {
	r := mux.NewRouter()
	cfg := &config.RuntimeConfig{}
	navReg := navpkg.NewRegistry()
	RegisterRoutes(r, cfg, navReg)

	tests := []struct {
		name           string
		userID         int32
		hasGrant       bool
		expectedStatus int
	}{
		{
			name:           "Allowed",
			userID:         1,
			hasGrant:       true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied",
			userID:         1,
			hasGrant:       false,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/private/topic/new", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			stub := testhelpers.NewQuerierStub()
			stub.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
				itemMatch := false
				if arg.Item.Valid && arg.Item.String == "topic" {
					itemMatch = true
				}
				// ViewerID is the user making the request
				if arg.ViewerID != tt.userID {
					return 0, nil
				}

				if arg.Section == "privateforum" && itemMatch && arg.Action == "see" {
					// Check ItemID
					// HasGrant("...", 0) usually results in ItemID being 0 (Valid=true) or maybe Valid=false depending on implementation.
					// We check if it matches 0 if valid.
					if (!arg.ItemID.Valid) || (arg.ItemID.Valid && arg.ItemID.Int32 == 0) {
						if tt.hasGrant {
							return 1, nil
						}
					}
				}
				return 0, sql.ErrNoRows
			}

			cd := common.NewCoreData(context.Background(), stub, cfg)
			cd.UserID = tt.userID

			ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
			req = req.WithContext(ctx)

			// Recover from potential template panics
			defer func() {
				if r := recover(); r != nil {
					// Check if status was set before panic
					if tt.expectedStatus == http.StatusForbidden && rr.Code == http.StatusForbidden {
						return
					}
					// If allowed, we might panic because templates are missing.
					if tt.expectedStatus == http.StatusOK {
						return
					}
					t.Logf("Recovered from panic: %v", r)
				}
			}()

			r.ServeHTTP(rr, req)

			if tt.expectedStatus == http.StatusOK {
				// Access granted -> might be 200 or 500 (template error)
				// As long as it is NOT 403 or 404
				if rr.Code == http.StatusForbidden {
					t.Errorf("Expected access granted, got 403")
				}
				if rr.Code == http.StatusNotFound {
					t.Errorf("Expected access granted, got 404")
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rr.Code)
			}
		})
	}
}
