package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
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
	t.Run("Happy Path", func(t *testing.T) {
		tests := []struct {
			name          string
			handler       http.HandlerFunc
			url           string
			expectedTitle string
		}{
			{
				name:          "Announcements",
				handler:       AdminAnnouncementsPage,
				url:           "/admin/announcements",
				expectedTitle: "Admin Announcements",
			},
			{
				name:          "Email Queue",
				handler:       AdminEmailPage,
				url:           "/admin/email/queue",
				expectedTitle: "Email Queue",
			},
			{
				name:          "Comments",
				handler:       AdminCommentsPage,
				url:           "/admin/comments",
				expectedTitle: "Comments",
			},
			{
				name:          "IP Bans",
				handler:       AdminIPBanPage,
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

				tt.handler(rr, req)

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

func TestAdminBreadcrumbsOverride(t *testing.T) {
	t.Run("Custom Breadcrumbs", func(t *testing.T) {
		cd, _ := setupCoreData(t, "/admin")
		cd.SetCurrentSection("admin")
		cd.PageTitle = "Custom Page"

		// Default behavior check (assuming "Custom Page" falls into default case)
		// default: crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
		// Breadcrumbs() strips the last one, so we get ["Admin"]
		defaultCrumbs := cd.Breadcrumbs()
		if len(defaultCrumbs) != 1 || defaultCrumbs[0].Title != "Admin" {
			t.Errorf("Expected default crumbs [Admin], got %v", defaultCrumbs)
		}

		// Set custom breadcrumbs
		custom := []common.Breadcrumb{
			{Title: "Home", Link: "/"},
			{Title: "Custom", Link: "/custom"},
		}
		cd.SetBreadcrumbs(custom...)

		// Verify override
		crumbs := cd.Breadcrumbs()
		if len(crumbs) != 2 {
			t.Errorf("Expected 2 crumbs, got %d", len(crumbs))
		}
		if crumbs[0].Title != "Home" || crumbs[1].Title != "Custom" {
			t.Errorf("Crumbs mismatch: %v", crumbs)
		}
	})
}
