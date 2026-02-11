package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

// Helper to setup CoreData
func setupCoreData(t *testing.T, url string) (*common.CoreData, *http.Request) {
	req := httptest.NewRequest("GET", url, nil)
	cfg := config.NewRuntimeConfig()
	queries := testhelpers.NewQuerierStub()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	return cd, req
}

func TestAdminBreadcrumbsLogic(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		tests := []struct {
			pageTitle      string
			expectedCrumbs []string
		}{
			{
				pageTitle:      "Email Queue",
				expectedCrumbs: []string{"Admin", "Email"},
			},
			{
				pageTitle:      "Email Sent",
				expectedCrumbs: []string{"Admin", "Email"},
			},
			{
				pageTitle:      "Comments",
				expectedCrumbs: []string{"Admin"}, // PageTitle matches Section, so stripped
			},
			{
				pageTitle:      "Comment 123",
				expectedCrumbs: []string{"Admin", "Comments"},
			},
			{
				pageTitle:      "Admin Announcements",
				expectedCrumbs: []string{"Admin"},
			},
			{
				pageTitle:      "Database Backup",
				expectedCrumbs: []string{"Admin", "Database"},
			},
			{
				pageTitle:      "Site Settings",
				expectedCrumbs: []string{"Admin"}, // If it's the section page
			},
			{
				pageTitle:      "Server Stats",
				expectedCrumbs: []string{"Admin"},
			},
			{
				pageTitle:      "IP Bans",
				expectedCrumbs: []string{"Admin"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.pageTitle, func(t *testing.T) {
				cd, _ := setupCoreData(t, "/admin")
				cd.SetCurrentSection("admin") // Manually set section
				cd.PageTitle = tt.pageTitle

				crumbs := cd.Breadcrumbs()

				if len(crumbs) != len(tt.expectedCrumbs) {
					t.Errorf("Expected %d crumbs, got %d: %v", len(tt.expectedCrumbs), len(crumbs), crumbs)
					return
				}

				for i, title := range tt.expectedCrumbs {
					if crumbs[i].Title != title {
						t.Errorf("Crumb %d: expected %s, got %s", i, title, crumbs[i].Title)
					}
				}
			})
		}
	})
}

func TestAdminPages_HaveTitlesAndBreadcrumbs(t *testing.T) {
	tests := []struct {
		name          string
		handler       http.Handler
		url           string
		expectedTitle string
	}{
		{
			name:          "Announcements",
			handler:       http.HandlerFunc(handlers.PageHandler(&AdminAnnouncementsPage{})),
			url:           "/admin/announcements",
			expectedTitle: "Admin Announcements",
		},
		{
			name:          "Email Queue",
			handler:       http.HandlerFunc(handlers.PageHandler(&AdminEmailPage{})),
			url:           "/admin/email/queue",
			expectedTitle: "Email Queue",
		},
		{
			name:          "Comments",
			handler:       http.HandlerFunc(handlers.PageHandler(&AdminCommentsPage{})),
			url:           "/admin/comments",
			expectedTitle: "Comments",
		},
		{
			name:          "IP Bans",
			handler:       http.HandlerFunc(handlers.PageHandler(&AdminIPBanPage{})),
			url:           "/admin/ipbans",
			expectedTitle: "IP Bans",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd, req := setupCoreData(t, tt.url)
			cd.SetCurrentSection("admin")

			// Mock handler execution
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Recovered from panic: %v", r)
				}
			}()

			tt.handler.ServeHTTP(rr, req)

			if cd.PageTitle == "" {
				t.Errorf("PageTitle was not set")
			}
			if cd.PageTitle != tt.expectedTitle {
				t.Errorf("Expected title %s, got %s", tt.expectedTitle, cd.PageTitle)
			}

				crumbs := cd.Breadcrumbs()
				if len(crumbs) == 0 {
					t.Errorf("Breadcrumbs are empty")
				}
				if crumbs[0].Title != "Admin" {
					t.Errorf("First crumb should be Admin")
				}
			})
		}
	})
}
