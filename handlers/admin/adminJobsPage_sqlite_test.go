//go:build sqlite

package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
)

func openAdminJobsTestDB(t *testing.T) (*sql.DB, db.Querier) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "admin_jobs.sqlite")
	sdb, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sdb.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sdb.Close() })
	schemaSQL, err := os.ReadFile(filepath.Join("..", "..", "database", "schema.sqlite.sql"))
	if err != nil {
		t.Fatalf("read schema.sqlite.sql: %v", err)
	}
	if _, err := sdb.ExecContext(context.Background(), string(schemaSQL)); err != nil {
		t.Fatalf("apply schema.sqlite.sql: %v", err)
	}
	return sdb, db.NewForDriver(sdb, "sqlite")
}

func TestAdminJobsPageAndActions(t *testing.T) {
	_, q := openAdminJobsTestDB(t)
	svc := jobs.NewConfiguredService(q)

	cd := common.NewCoreData(
		context.Background(),
		q,
		config.NewRuntimeConfig(),
		common.WithUserID(1),
		common.WithJobService(svc),
	)

	// 1. Disallowed/loopback URL must be rejected without creating a durable job.
	form := url.Values{"action": {"enqueue_url"}, "url": {"http://127.0.0.1/internal"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/jobs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()
	AdminJobsAction(rec, req)
	if rec.Code != http.StatusSeeOther || !strings.Contains(rec.Header().Get("Location"), "error=") {
		t.Fatalf("loopback URL enqueue status=%d loc=%q, want redirect with error", rec.Code, rec.Header().Get("Location"))
	}

	// 2. Valid external URL is enqueued durably.
	form = url.Values{"action": {"enqueue_url"}, "url": {"https://example.com/post"}}
	req = httptest.NewRequest(http.MethodPost, "/admin/jobs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec = httptest.NewRecorder()
	AdminJobsAction(rec, req)
	if rec.Code != http.StatusSeeOther || strings.Contains(rec.Header().Get("Location"), "error=") {
		t.Fatalf("valid URL enqueue status=%d loc=%q", rec.Code, rec.Header().Get("Location"))
	}

	list, err := svc.AdminList(context.Background(), "", "", 10, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("AdminList = (%d, %v), want 1", len(list), err)
	}
	jobID := list[0].ID

	// 3. Cancel pending job -> status becomes cancelled.
	form = url.Values{"action": {"cancel"}, "id": {"1"}}
	req = httptest.NewRequest(http.MethodPost, "/admin/jobs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec = httptest.NewRecorder()
	AdminJobsAction(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("cancel status=%d", rec.Code)
	}
	cancelled, _ := q.SystemGetDurableJobByID(context.Background(), jobID)
	if cancelled.Status != jobs.StatusCancelled {
		t.Fatalf("after cancel status=%q, want cancelled", cancelled.Status)
	}

	// 4. Retry cancelled job -> status returns to pending.
	form = url.Values{"action": {"retry"}, "id": {"1"}}
	req = httptest.NewRequest(http.MethodPost, "/admin/jobs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec = httptest.NewRecorder()
	AdminJobsAction(rec, req)
	retried, _ := q.SystemGetDurableJobByID(context.Background(), jobID)
	if retried.Status != jobs.StatusPending {
		t.Fatalf("after retry status=%q, want pending", retried.Status)
	}

	// 5. Cancel and purge -> terminal row removed.
	_, _ = svc.AdminCancel(context.Background(), jobID)
	form = url.Values{"action": {"purge"}, "status": {"cancelled"}}
	req = httptest.NewRequest(http.MethodPost, "/admin/jobs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec = httptest.NewRecorder()
	AdminJobsAction(rec, req)
	afterPurge, _ := svc.AdminList(context.Background(), "", "", 10, 0)
	if len(afterPurge) != 0 {
		t.Fatalf("after purge len=%d, want 0", len(afterPurge))
	}
}
