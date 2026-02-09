package blogs

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/handlertest"
	"github.com/arran4/goa4web/internal/db"
)

func TestBlogAddPage_AccessControl(t *testing.T) {
	tests := []struct {
		name           string
		userID         int32
		isAdmin        bool
		grants         []*db.Grant
		expectedStatus int
	}{
		{
			name:           "NoUser_Forbidden",
			userID:         0,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "UserNoGrant_Forbidden",
			userID:         123,
			grants:         []*db.Grant{},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "UserWithGrant_Allowed",
			userID: 123,
			grants: []*db.Grant{
				{
					Section: "blogs",
					Item:    sql.NullString{String: "entry", Valid: true},
					Action:  "post",
					Active:  true,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Admin_Allowed",
			userID:         123,
			isAdmin:        true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req := httptest.NewRequest("GET", "/add", nil)
			w := httptest.NewRecorder()

			// Prepare CoreData options
			var opts []common.CoreOption

			if tt.isAdmin {
				// Mock admin permissions (used by HasAdminRole)
				opts = append(opts, common.WithPermissions([]*db.GetPermissionsByUserIDRow{
					{IsAdmin: true, Name: "admin"},
				}))
			}

			// Initialize CoreData
			req, cd, stub := handlertest.RequestWithCoreData(t, req, opts...)
			cd.UserID = tt.userID
			if tt.isAdmin {
				cd.AdminMode = true // Force admin mode for test
			}

			// Configure stub grants
			grantMap := make(map[string]bool)
			for _, g := range tt.grants {
				key := g.Section + "|"
				if g.Item.Valid {
					key += g.Item.String
				}
				key += "|" + g.Action
				grantMap[key] = true
			}

			stub.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
				item := ""
				if arg.Item.Valid {
					item = arg.Item.String
				}
				key := arg.Section + "|" + item + "|" + arg.Action
				if grantMap[key] {
					return 1, nil
				}
				return 0, sql.ErrNoRows
			}

			// Wrap the handler with EnforceGrant middleware
			handler := handlers.EnforceGrant(addBlogTask.Page, nil, "blogs", "entry", "post", 0)

			// Execute
			handler.ServeHTTP(w, req)

			// Check status
			if tt.expectedStatus == http.StatusForbidden {
				if w.Code != http.StatusForbidden {
					t.Errorf("expected status 403, got %d", w.Code)
				}
			} else {
				// We expect 200 OK or 500 Internal Server Error (template missing).
				// But definitively NOT 403.
				if w.Code == http.StatusForbidden {
					t.Errorf("expected access allowed, but got 403 Forbidden")
				}
			}
		})
	}
}
