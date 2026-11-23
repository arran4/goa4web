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
	"github.com/arran4/goa4web/internal/router"
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
	sdb, q := openAdminJobsTestDB(t)
	svc := jobs.NewConfiguredService(q)

	cd := common.NewCoreData(
		context.Background(),
		q,
		config.NewRuntimeConfig(),
		common.WithUserID(1),
		common.WithUserRoles([]string{"administrator"}),
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

	// 2. Valid unused external URL creates an external_links record and durable job.
	form = url.Values{"action": {"enqueue_url"}, "url": {"https://example.com/post"}}
	req = httptest.NewRequest(http.MethodPost, "/admin/jobs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec = httptest.NewRecorder()
	AdminJobsAction(rec, req)
	if rec.Code != http.StatusSeeOther || strings.Contains(rec.Header().Get("Location"), "error=") {
		t.Fatalf("valid URL enqueue status=%d loc=%q", rec.Code, rec.Header().Get("Location"))
	}

	linkRow, err := q.GetExternalLink(context.Background(), "https://example.com/post")
	if err != nil || linkRow == nil || linkRow.ID <= 0 {
		t.Fatalf("GetExternalLink = (%#v, %v), want created external_links row", linkRow, err)
	}

	list, err := svc.AdminList(context.Background(), "", "", 10, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("AdminList = (%d, %v), want 1", len(list), err)
	}
	if list[0].TargetID != int64(linkRow.ID) {
		t.Fatalf("job TargetID = %d, want external_links ID %d", list[0].TargetID, linkRow.ID)
	}
	jobID := list[0].ID

	// Re-enqueueing the same URL reuses the existing external_links record.
	req = httptest.NewRequest(http.MethodPost, "/admin/jobs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec = httptest.NewRecorder()
	AdminJobsAction(rec, req)
	if rec.Code != http.StatusSeeOther || strings.Contains(rec.Header().Get("Location"), "error=") {
		t.Fatalf("duplicate URL enqueue status=%d loc=%q", rec.Code, rec.Header().Get("Location"))
	}
	var linkCount int
	if err := sdb.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM external_links WHERE url = 'https://example.com/post'`).Scan(&linkCount); err != nil || linkCount != 1 {
		t.Fatalf("external_links count = (%d, %v), want 1", linkCount, err)
	}

	// Render GET /admin/jobs?id=<jobID> and verify Next Run and detail metadata.
	getReq := httptest.NewRequest(http.MethodGet, "/admin/jobs?id=1", nil)
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), consts.KeyCoreData, cd))
	getRec := httptest.NewRecorder()
	AdminJobsPage(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("AdminJobsPage GET status=%d", getRec.Code)
	}
	body := getRec.Body.String()
	for _, want := range []string{"Next Run", "Job #1 Details", "https://example.com/post"} {
		if !strings.Contains(body, want) {
			t.Fatalf("AdminJobsPage body missing %q", want)
		}
	}

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

func TestAdminJobsRoutesRejectUnauthenticatedAndNonAdmin(t *testing.T) {
	sdb, q := openAdminJobsTestDB(t)
	svc := jobs.NewConfiguredService(q)

	getHandler := router.AdminCheckerMiddleware(http.HandlerFunc(AdminJobsPage))
	postHandler := router.AdminCheckerMiddleware(http.HandlerFunc(AdminJobsAction))

	for _, tc := range []struct {
		name  string
		uid   int32
		roles []string
	}{
		{name: "anonymous", uid: 0, roles: nil},
		{name: "non-admin-member", uid: 42, roles: []string{"member"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cd := common.NewCoreData(
				context.Background(),
				q,
				config.NewRuntimeConfig(),
				common.WithUserID(tc.uid),
				common.WithUserRoles(tc.roles),
				common.WithJobService(svc),
			)

			getReq := httptest.NewRequest(http.MethodGet, "/admin/jobs", nil)
			getReq = getReq.WithContext(context.WithValue(getReq.Context(), consts.KeyCoreData, cd))
			getRec := httptest.NewRecorder()
			getHandler.ServeHTTP(getRec, getReq)
			if getRec.Code != http.StatusForbidden {
				t.Fatalf("GET /admin/jobs status=%d, want %d", getRec.Code, http.StatusForbidden)
			}

			form := url.Values{"action": {"enqueue_url"}, "url": {"https://example.com/blocked"}}
			postReq := httptest.NewRequest(http.MethodPost, "/admin/jobs", strings.NewReader(form.Encode()))
			postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			postReq = postReq.WithContext(context.WithValue(postReq.Context(), consts.KeyCoreData, cd))
			postRec := httptest.NewRecorder()
			postHandler.ServeHTTP(postRec, postReq)
			if postRec.Code != http.StatusForbidden {
				t.Fatalf("POST /admin/jobs status=%d, want %d", postRec.Code, http.StatusForbidden)
			}
		})
	}

	var jobCount, linkCount int
	if err := sdb.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM durable_jobs`).Scan(&jobCount); err != nil || jobCount != 0 {
		t.Fatalf("durable_jobs count = (%d, %v), want 0", jobCount, err)
	}
	if err := sdb.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM external_links`).Scan(&linkCount); err != nil || linkCount != 0 {
		t.Fatalf("external_links count = (%d, %v), want 0", linkCount, err)
	}
}
