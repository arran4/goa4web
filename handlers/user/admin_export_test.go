package user

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type mockQuerier struct {
	db.Querier
	user *db.SystemGetUserByIDRow
}

func (m *mockQuerier) SystemGetUserByID(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if m.user != nil && m.user.Idusers == id {
		return m.user, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockQuerier) GetPreferenceForLister(ctx context.Context, id int32) (*db.Preference, error) {
	return &db.Preference{
		UsersIdusers: id,
	}, nil
}

func (m *mockQuerier) GetUserLanguages(ctx context.Context, id int32) ([]*db.UserLanguage, error) {
	return []*db.UserLanguage{}, nil
}

func (m *mockQuerier) GetPermissionsByUserID(ctx context.Context, id int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return []*db.GetPermissionsByUserIDRow{}, nil
}

func (m *mockQuerier) SystemListWritingCategories(ctx context.Context, arg db.SystemListWritingCategoriesParams) ([]*db.WritingCategory, error) {
	return []*db.WritingCategory{}, nil
}

func (m *mockQuerier) AdminGetAllWritingsByAuthor(ctx context.Context, authorID int32) ([]*db.AdminGetAllWritingsByAuthorRow, error) {
	return []*db.AdminGetAllWritingsByAuthorRow{}, nil
}

func (m *mockQuerier) AdminGetAllBlogEntriesByUser(ctx context.Context, authorID int32) ([]*db.AdminGetAllBlogEntriesByUserRow, error) {
	return []*db.AdminGetAllBlogEntriesByUserRow{}, nil
}

func (m *mockQuerier) AdminGetThreadsStartedByUser(ctx context.Context, usersIdusers int32) ([]*db.Forumthread, error) {
	return []*db.Forumthread{}, nil
}

func (m *mockQuerier) AdminGetAllCommentsByUser(ctx context.Context, usersIdusers int32) ([]*db.AdminGetAllCommentsByUserRow, error) {
	return []*db.AdminGetAllCommentsByUserRow{}, nil
}

func TestAdminUsersExportPage(t *testing.T) {
	ctx := context.Background()
	queries := &mockQuerier{
		user: &db.SystemGetUserByIDRow{
			Idusers:  123,
			Username: sql.NullString{String: "testuser", Valid: true},
		},
	}
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(ctx, queries, cfg)
	cd.UserID = 999 // Admin user

	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req, err := http.NewRequestWithContext(ctx, "GET", "/admin/users/export?uid=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	adminUsersExportPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/zip" {
		t.Errorf("expected Content-Type application/zip, got %v", contentType)
	}

	contentDisposition := rec.Header().Get("Content-Disposition")
	if contentDisposition != "attachment; filename=user_123.zip" {
		t.Errorf("expected Content-Disposition attachment; filename=user_123.zip, got %v", contentDisposition)
	}

	body := rec.Body.Bytes()
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("failed to read zip: %v", err)
	}

	foundUserJson := false
	for _, f := range zr.File {
		if f.Name == "user.json" {
			foundUserJson = true
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open user.json: %v", err)
			}
			var data struct {
				User *db.User `json:"user"`
			}
			if err := json.NewDecoder(rc).Decode(&data); err != nil {
				t.Fatalf("failed to decode user.json: %v", err)
			}
			rc.Close()

			if data.User.Idusers != 123 {
				t.Errorf("expected user id 123, got %d", data.User.Idusers)
			}
			if data.User.Username.String != "testuser" {
				t.Errorf("expected username testuser, got %s", data.User.Username.String)
			}
			// Verify PublicProfileEnabledAt is preserved (nil in this case, but field should exist)
			// Since we mock it as default, it's nil.
		}
	}
	if !foundUserJson {
		t.Error("user.json not found in zip")
	}
}
