package privateforum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestPage_NoAccess(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	Page(w, req)

	if body := w.Body.String(); !strings.Contains(body, "may not have permission") {
		t.Fatalf("expected no access message, got %q", body)
	}
}

func TestPage_Access(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	Page(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "Private Topics") {
		t.Fatalf("expected private topics page, got %q", body)
	}
	if !strings.Contains(body, "<form id=\"private-form\"") {
		t.Fatalf("expected create form, got %q", body)
	}
}

type querierMock struct {
	db.Querier
	hasGrant func(ctx context.Context, arg db.HasGrantParams) (int32, error)
}

func (q *querierMock) HasGrant(ctx context.Context, arg db.HasGrantParams) (int32, error) {
	return q.hasGrant(ctx, arg)
}

func TestPage_SeeNoCreate(t *testing.T) {
	callCount := 0
	mockQueries := &querierMock{
		hasGrant: func(ctx context.Context, arg db.HasGrantParams) (int32, error) {
			callCount++
			if callCount == 1 {
				// First call, permission granted
				return 1, nil
			}
			// Second call, permission denied
			return 0, sql.ErrNoRows
		},
	}
	cd := common.NewCoreData(context.Background(), mockQueries, config.NewRuntimeConfig())
	cd.UserID = 1

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	Page(w, req)

	body := w.Body.String()
	if strings.Contains(body, "Start conversation") {
		t.Fatalf("unexpected create form, got %q", body)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 calls to HasGrant, got %d", callCount)
	}
}
