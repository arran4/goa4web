package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type filesQueries struct {
	db.Querier
	images []*db.AdminListAllImagePostsRow
	count  int64
	// map path -> row for GetImagePostInfoByPath
	managedFiles map[string]*db.GetImagePostInfoByPathRow
}

func (q *filesQueries) AdminListAllImagePosts(ctx context.Context, arg db.AdminListAllImagePostsParams) ([]*db.AdminListAllImagePostsRow, error) {
	return q.images, nil
}

func (q *filesQueries) AdminCountAllImagePosts(ctx context.Context) (int64, error) {
	return q.count, nil
}

func (q *filesQueries) GetImagePostInfoByPath(ctx context.Context, arg db.GetImagePostInfoByPathParams) (*db.GetImagePostInfoByPathRow, error) {
	if q.managedFiles != nil {
		if row, ok := q.managedFiles[arg.Fullimage.String]; ok {
			return row, nil
		}
	}
	return nil, sql.ErrNoRows
}

func TestAdminFilesPage(t *testing.T) {
	now := time.Now()
	queries := &filesQueries{
		count: 1,
		images: []*db.AdminListAllImagePostsRow{{
			Idimagepost: 123,
			Fullimage:   sql.NullString{Valid: true, String: "/imagebbs/images/test.jpg"},
			Description: sql.NullString{Valid: true, String: "Test Image"},
			Posted:      sql.NullTime{Valid: true, Time: now},
		}},
	}

	req := httptest.NewRequest("GET", "/admin/files", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AdminFilesPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Managed Files") {
		t.Fatalf("missing title")
	}
	if !strings.Contains(body, "/imagebbs/images/test.jpg") {
		t.Fatalf("missing image path: %s", body)
	}
	if !strings.Contains(body, "View Unmanaged Files") {
		t.Fatalf("missing unmanaged link: %s", body)
	}
}

func TestAdminUnmanagedFilesPage(t *testing.T) {
	// Create temp dir
	tmpDir := t.TempDir()

	// Create unmanaged file
	if err := os.WriteFile(filepath.Join(tmpDir, "unmanaged.jpg"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	// Create managed file
	if err := os.WriteFile(filepath.Join(tmpDir, "managed.jpg"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	queries := &filesQueries{
		managedFiles: map[string]*db.GetImagePostInfoByPathRow{
			"/imagebbs/images/managed.jpg": {
				Idimagepost: 1,
				Title:       sql.NullString{Valid: true, String: "Board"},
				Username:    sql.NullString{Valid: true, String: "User"},
			},
		},
	}

	req := httptest.NewRequest("GET", "/admin/files/unmanaged", nil)
	ctx := req.Context()
	cfg := config.NewRuntimeConfig()
	cfg.ImageUploadDir = tmpDir
	cd := common.NewCoreData(ctx, queries, cfg, common.WithUserRoles([]string{"administrator"}))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AdminUnmanagedFilesPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	// fmt.Println(body)
	if !strings.Contains(body, "Unmanaged Files") {
		t.Fatalf("missing title")
	}
	if !strings.Contains(body, "unmanaged.jpg") {
		t.Fatalf("missing unmanaged file")
	}
	if strings.Contains(body, "<td>managed.jpg</td>") {
		t.Fatalf("managed file should be filtered out")
	}
	if !strings.Contains(body, "Delete") {
		t.Fatalf("missing delete button")
	}
}
