package goa4web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBlogsBlogAddPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), ContextValues("coreData"), &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	blogsBlogAddPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestBlogsBlogEditPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/1/edit", nil)
	ctx := context.WithValue(req.Context(), ContextValues("coreData"), &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	blogsBlogEditPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestGetPermissionsByUserIdAndSectionBlogsPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin/blogs/user/permissions", nil)
	ctx := context.WithValue(req.Context(), ContextValues("coreData"), &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	getPermissionsByUserIdAndSectionBlogsPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}
