package imagebbs

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestRequireImagebbsGrantWithBoard(t *testing.T) {
	allow := true
	queries := testhelpers.NewQuerierStub()
	queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		if allow {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 99

	req := httptest.NewRequest("GET", "/admin/imagebbs/board/7", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"board": "7"})

	if !requireImagebbsGrant(imagebbsApproveAction)(req, &mux.RouteMatch{}) {
		t.Fatalf("expected matcher to allow request with grant")
	}
	if len(queries.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one grant check, got %d", len(queries.SystemCheckGrantCalls))
	}
	want := db.SystemCheckGrantParams{
		ViewerID: cd.UserID,
		Section:  "imagebbs",
		Item:     sql.NullString{String: "board", Valid: true},
		Action:   imagebbsApproveAction,
		ItemID:   sql.NullInt32{Int32: 7, Valid: true},
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: true},
	}
	if got := queries.SystemCheckGrantCalls[0]; got != want {
		t.Fatalf("expected grant call %#v, got %#v", want, got)
	}
}

func TestRequireImagebbsGrantWithPost(t *testing.T) {
	allow := true
	queries := testhelpers.NewQuerierStub()
	queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		if allow {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}
	queries.AdminGetImagePostRow = &db.AdminGetImagePostRow{
		Idimagepost:            5,
		ImageboardIdimageboard: sql.NullInt32{Int32: 3, Valid: true},
	}
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 42

	req := httptest.NewRequest("GET", "/admin/imagebbs/approve/5", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"post": "5"})

	if !requireImagebbsGrant(imagebbsApproveAction)(req, &mux.RouteMatch{}) {
		t.Fatalf("expected matcher to allow request with post-derived grant")
	}
	if len(queries.AdminGetImagePostCalls) != 1 {
		t.Fatalf("expected one image post lookup, got %d", len(queries.AdminGetImagePostCalls))
	}
	if got := queries.AdminGetImagePostCalls[0]; got != 5 {
		t.Fatalf("expected image post lookup for 5, got %d", got)
	}
	if len(queries.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one grant check, got %d", len(queries.SystemCheckGrantCalls))
	}
	want := db.SystemCheckGrantParams{
		ViewerID: cd.UserID,
		Section:  "imagebbs",
		Item:     sql.NullString{String: "board", Valid: true},
		Action:   imagebbsApproveAction,
		ItemID:   sql.NullInt32{Int32: 3, Valid: true},
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: true},
	}
	if got := queries.SystemCheckGrantCalls[0]; got != want {
		t.Fatalf("expected grant call %#v, got %#v", want, got)
	}
}

func TestApprovePostTaskDeniesWithoutGrant(t *testing.T) {
	allow := false
	queries := testhelpers.NewQuerierStub()
	queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		if allow {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}
	queries.AdminGetImagePostRow = &db.AdminGetImagePostRow{
		Idimagepost:            12,
		ImageboardIdimageboard: sql.NullInt32{Int32: 9, Valid: true},
	}
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 77

	req := httptest.NewRequest("POST", "/admin/imagebbs/approve/12", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"post": "12"})

	if _, ok := approvePostTask.Action(httptest.NewRecorder(), req).(http.HandlerFunc); !ok {
		t.Fatalf("expected forbidden handler when grant is missing")
	}
	if len(queries.AdminApproveImagePostCalls) != 0 {
		t.Fatalf("expected no approval calls, got %d", len(queries.AdminApproveImagePostCalls))
	}
}

func TestApprovePostTaskAllowsWithGrant(t *testing.T) {
	allow := true
	queries := testhelpers.NewQuerierStub()
	queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		if allow {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}
	queries.AdminGetImagePostRow = &db.AdminGetImagePostRow{
		Idimagepost:            4,
		ImageboardIdimageboard: sql.NullInt32{Int32: 2, Valid: true},
	}
	queries.AdminApproveImagePostFn = func(context.Context, int32) error {
		return nil
	}
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 15

	req := httptest.NewRequest("POST", "/admin/imagebbs/approve/4", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"post": "4"})

	if result := approvePostTask.Action(httptest.NewRecorder(), req); result != nil {
		t.Fatalf("unexpected result %v", result)
	}
	if len(queries.AdminApproveImagePostCalls) != 1 {
		t.Fatalf("expected one approval call, got %d", len(queries.AdminApproveImagePostCalls))
	}
	if got := queries.AdminApproveImagePostCalls[0]; got != 4 {
		t.Fatalf("expected approval call for 4, got %d", got)
	}
}
