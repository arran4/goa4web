package common_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestTemplateFuncsFirstline(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	first := funcs["firstline"].(func(string) string)
	if got := first("a\nb\n"); got != "a" {
		t.Errorf("firstline=%q", got)
	}
}

func TestTemplateFuncsLeft(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	left := funcs["left"].(func(int, string) string)
	if got := left(3, "hello"); got != "hel" {
		t.Errorf("left short=%q", got)
	}
	if got := left(10, "hi"); got != "hi" {
		t.Errorf("left long=%q", got)
	}
}

func TestTemplateFuncsCSRFToken(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	if _, ok := funcs["csrfToken"]; !ok {
		t.Errorf("csrfToken func missing")
	}
	if _, ok := funcs["csrf"]; ok {
		t.Errorf("csrf func should not be present")
	}
}

func TestLatestNewsRespectsPermissions(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	var grantChecks []int32
	fetchCalls := 0
	cfg := config.NewRuntimeConfig()
	pageSize := cfg.PageSizeDefault
	cd := common.NewCoreData(req.Context(), nil, cfg,
		common.WithUserRoles([]string{"user"}),
		common.WithGrantChecker(func(section, item, action string, itemID int32) bool {
			grantChecks = append(grantChecks, itemID)
			return itemID == 1
		}),
		common.WithNewsFetcher(func(offset, limit int32) ([]*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
			fetchCalls++
			if offset != 0 {
				t.Fatalf("unexpected offset %d", offset)
			}
			if limit != int32(pageSize) {
				t.Fatalf("unexpected limit %d", limit)
			}
			return []*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow{
				{Idsitenews: 1},
				{Idsitenews: 2},
			}, nil
		}))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	_ = req.WithContext(ctx)

	res, err := cd.LatestNews()
	if err != nil {
		t.Fatalf("LatestNews: %v", err)
	}
	if l := len(res); l != 1 {
		t.Fatalf("expected 1 news post, got %d", l)
	}
	if res[0].Idsitenews != 1 {
		t.Fatalf("unexpected news id %d", res[0].Idsitenews)
	}
	if len(grantChecks) != 2 || grantChecks[0] != 1 || grantChecks[1] != 2 {
		t.Fatalf("grant checks = %v, want [1 2]", grantChecks)
	}
	if fetchCalls != 1 {
		t.Fatalf("expected one fetch call, got %d", fetchCalls)
	}
}

func TestAddmodeSkipsAdminLinks(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{AdminMode: true}
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	funcs := cd.Funcs(req)
	addmode := funcs["addmode"].(func(string) string)

	tests := []struct {
		in   string
		want string
	}{
		{"/admin", "/admin"},
		{"/admin/tools", "/admin/tools"},
		{"/admin/tools?flag=1", "/admin/tools?flag=1"},
		{"http://example.com/admin", "http://example.com/admin"},
		{"/administrator", "/administrator?mode=admin"},
		{"/user", "/user?mode=admin"},
		{"/user?id=1", "/user?id=1&mode=admin"},
	}

	for _, tt := range tests {
		if got := addmode(tt.in); got != tt.want {
			t.Errorf("addmode(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
