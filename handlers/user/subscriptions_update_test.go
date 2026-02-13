package user

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUpdateSubscriptionsTask_MandatoryProtection(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		// Setup
		mandatoryPattern := "password reset:/auth/reset" // Defined as mandatory in definitions.go
		uid := int32(123)

		tests := []struct {
			name           string
			existing       []*db.ListSubscriptionsByUserRow
			presented      []string // "pattern|method"
			subs           []string // "pattern|method" (checked ones)
			expectDelete   bool
			expectInsert   bool
			expectedDelete string
			isAdmin        bool
		}{
			{
				name: "Mandatory sub removal attempt",
				existing: []*db.ListSubscriptionsByUserRow{
					{Pattern: mandatoryPattern, Method: "internal"},
				},
				presented:    []string{mandatoryPattern + "|internal"},
				subs:         []string{}, // Empty, implying user unchecked it
				expectDelete: false,      // Should NOT delete because it's mandatory
			},
			{
				name: "Normal sub removal",
				existing: []*db.ListSubscriptionsByUserRow{
					{Pattern: "post:/blog/*", Method: "internal"},
				},
				presented:      []string{"post:/blog/*|internal"},
				subs:           []string{},
				expectDelete:   true,
				expectedDelete: "post:/blog/*",
			},
			{
				name:         "Add admin sub - Non-Admin",
				existing:     []*db.ListSubscriptionsByUserRow{},
				presented:    []string{"notify:/admin/*|internal"},
				subs:         []string{"notify:/admin/*|internal"},
				expectInsert: false, // Security check should prevent it
				isAdmin:      false,
			},
			{
				name:         "Add admin sub - Admin",
				existing:     []*db.ListSubscriptionsByUserRow{},
				presented:    []string{"notify:/admin/*|internal"},
				subs:         []string{"notify:/admin/*|internal"},
				expectInsert: true,
				isAdmin:      true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := testhelpers.NewQuerierStub()
				q.ListSubscriptionsByUserReturns = tt.existing
				q.GetPermissionsByUserIDFn = func(id int32) ([]*db.GetPermissionsByUserIDRow, error) {
					if tt.isAdmin {
						return []*db.GetPermissionsByUserIDRow{{Name: "admin", IsAdmin: true}}, nil
					}
					return []*db.GetPermissionsByUserIDRow{}, nil
				}

				// Mock Session
				mockSession := &sessions.Session{
					Values: map[interface{}]interface{}{"UID": uid},
				}

				cd := common.NewCoreData(context.Background(), q, nil, common.WithSession(mockSession))
				cd.UserID = uid
				if tt.isAdmin {
					cd.AdminMode = true
				}

				// Build request
				form := url.Values{}
				for _, p := range tt.presented {
					form.Add("presented_subs", p)
				}
				for _, s := range tt.subs {
					form.Add("subs", s)
				}

				req := httptest.NewRequest("POST", "/usr/subscriptions/update", nil)
				req.PostForm = form
				ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
				ctx = context.WithValue(ctx, core.ContextValues("session"), mockSession)
				req = req.WithContext(ctx)
				w := httptest.NewRecorder()

				// Execute
				updateSubscriptionsTask.Action(w, req)

				// Verify Delete
				if tt.expectDelete {
					found := false
					for _, call := range q.DeleteSubscriptionParams {
						if call.Pattern == tt.expectedDelete {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected delete call for %s, but not found", tt.expectedDelete)
					}
				} else {
					if len(q.DeleteSubscriptionParams) > 0 {
						t.Errorf("Unexpected delete calls: %v", q.DeleteSubscriptionParams)
					}
				}

				// Verify Insert
				if tt.expectInsert {
					if len(q.InsertSubscriptionParams) == 0 {
						t.Errorf("Expected insert call, but not found")
					}
				} else {
					// Only fail if we didn't expect insert but got one.
					// However, "Normal sub removal" case shouldn't insert.
					// "Mandatory sub removal" case shouldn't insert (it's already there).
					if len(q.InsertSubscriptionParams) > 0 {
						t.Errorf("Unexpected insert calls: %v", q.InsertSubscriptionParams)
					}
				}
			})
		}
	})
}
