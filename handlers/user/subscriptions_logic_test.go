package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/subscriptions"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestUserSubscriptionsPage_AdminOptionsVisibility(t *testing.T) {
	// Save original Handle and restore it after test
	originalHandle := tasks.Handle
	defer func() { tasks.Handle = originalHandle }()

	var capturedData any
	tasks.Handle = func(w http.ResponseWriter, r *http.Request, tmpl tasks.Template, data any) error {
		capturedData = data
		return nil
	}

	tests := []struct {
		name           string
		adminMode      bool
		queryMode      string
		hasAdminRole   bool
		expectAdminOps bool
	}{
		{
			name:           "Non-Admin User, AdminMode=False",
			adminMode:      false,
			hasAdminRole:   false,
			expectAdminOps: false,
		},
		{
			name:           "Non-Admin User, AdminMode=True (Exploit Attempt)",
			adminMode:      true,
			hasAdminRole:   false,
			expectAdminOps: false,
		},
		{
			name:           "Admin User, AdminMode=False",
			adminMode:      false,
			hasAdminRole:   true,
			expectAdminOps: false,
		},
		{
			name:           "Admin User, AdminMode=True",
			adminMode:      true,
			hasAdminRole:   true,
			expectAdminOps: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedData = nil // Reset captured data

			// Setup QuerierStub
			q := &db.QuerierStub{
				ListSubscriptionsByUserReturns: []*db.ListSubscriptionsByUserRow{},
			}
			if tt.hasAdminRole {
				q.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{
					{Name: "admin", IsAdmin: true},
				}
			} else {
				q.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{}
			}

			// Setup CoreData
			cd := common.NewCoreData(context.Background(), q, nil)
			cd.UserID = 1
			cd.AdminMode = tt.adminMode

			// Create request with CoreData
			url := "/usr/subscriptions"
			if tt.queryMode != "" {
				url += "?mode=" + tt.queryMode
			}
			req := httptest.NewRequest("GET", url, nil)
			ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// Call the handler
			userSubscriptionsPage(w, req)

			// Verify results
			if capturedData == nil {
				t.Fatal("Handle was not called or data was nil")
			}

			dataStruct, ok := capturedData.(struct {
				Groups      []*subscriptions.SubscriptionGroup
				AdminGroups []*subscriptions.SubscriptionGroup
				IsAdminMode bool
			})
			if !ok {
				t.Fatalf("Data is not the expected struct type, got %T", capturedData)
			}

			// Check if AdminGroups has content
			hasAdminGroups := len(dataStruct.AdminGroups) > 0
			if hasAdminGroups != tt.expectAdminOps {
				t.Errorf("Expected AdminGroups presence to be %v, got %v", tt.expectAdminOps, hasAdminGroups)
			}

			// Ensure Groups (regular) does NOT have admin ops
			for _, g := range dataStruct.Groups {
				if g.Definition.IsAdminOnly {
					t.Errorf("Found admin subscription %s in regular Groups", g.Name)
				}
			}
		})
	}
}
