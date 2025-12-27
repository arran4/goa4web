package imagebbs

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestRequireImagebbsGrantWithBoard(t *testing.T) {
	stub := &db.QuerierStub{}
	cd := common.NewCoreData(context.Background(), stub, config.NewRuntimeConfig())
	cd.UserID = 99

	stub.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		if arg.ViewerID != cd.UserID {
			t.Fatalf("unexpected viewer id %d", arg.ViewerID)
		}
		if arg.Section != "imagebbs" || arg.Item.String != "board" || arg.Action != imagebbsApproveAction {
			t.Fatalf("unexpected grant args: %#v", arg)
		}
		if arg.ItemID != (sql.NullInt32{Int32: 7, Valid: true}) {
			t.Fatalf("unexpected item id: %#v", arg.ItemID)
		}
		if arg.UserID != (sql.NullInt32{Int32: cd.UserID, Valid: true}) {
			t.Fatalf("unexpected user id: %#v", arg.UserID)
		}
		return 1, nil
	}

	req := httptest.NewRequest("GET", "/admin/imagebbs/board/7", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"board": "7"})

	if !requireImagebbsGrant(imagebbsApproveAction)(req, &mux.RouteMatch{}) {
		t.Fatalf("expected matcher to allow request with grant")
	}
}

func TestRequireImagebbsGrantWithPost(t *testing.T) {
	stub := &db.QuerierStub{
		AdminGetImagePostRow: &db.AdminGetImagePostRow{
			Idimagepost:            5,
			ImageboardIdimageboard: sql.NullInt32{Int32: 3, Valid: true},
		},
	}
	cd := common.NewCoreData(context.Background(), stub, config.NewRuntimeConfig())
	cd.UserID = 42

	stub.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		if arg.ItemID != (sql.NullInt32{Int32: 3, Valid: true}) {
			t.Fatalf("unexpected item id: %#v", arg.ItemID)
		}
		if arg.Action != imagebbsApproveAction {
			t.Fatalf("unexpected action %s", arg.Action)
		}
		return 1, nil
	}

	req := httptest.NewRequest("GET", "/admin/imagebbs/approve/5", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"post": "5"})

	if !requireImagebbsGrant(imagebbsApproveAction)(req, &mux.RouteMatch{}) {
		t.Fatalf("expected matcher to allow request with post-derived grant")
	}
	if len(stub.AdminGetImagePostCalls) != 1 || stub.AdminGetImagePostCalls[0] != 5 {
		t.Fatalf("unexpected AdminGetImagePost calls: %#v", stub.AdminGetImagePostCalls)
	}
}

func TestApprovePostTaskDeniesWithoutGrant(t *testing.T) {
	stub := &db.QuerierStub{
		AdminGetImagePostRow: &db.AdminGetImagePostRow{
			Idimagepost:            12,
			ImageboardIdimageboard: sql.NullInt32{Int32: 9, Valid: true},
		},
		SystemCheckGrantErr: errors.New("denied"),
	}
	cd := common.NewCoreData(context.Background(), stub, config.NewRuntimeConfig())
	cd.UserID = 77

	req := httptest.NewRequest("POST", "/admin/imagebbs/approve/12", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"post": "12"})

	if _, ok := approvePostTask.Action(httptest.NewRecorder(), req).(http.HandlerFunc); !ok {
		t.Fatalf("expected forbidden handler when grant is missing")
	}
	if len(stub.AdminApproveImagePostCalls) != 0 {
		t.Fatalf("unexpected approve calls: %#v", stub.AdminApproveImagePostCalls)
	}
}

func TestApprovePostTaskAllowsWithGrant(t *testing.T) {
	stub := &db.QuerierStub{
		AdminGetImagePostRow: &db.AdminGetImagePostRow{
			Idimagepost:            4,
			ImageboardIdimageboard: sql.NullInt32{Int32: 2, Valid: true},
		},
	}
	cd := common.NewCoreData(context.Background(), stub, config.NewRuntimeConfig())
	cd.UserID = 15

	req := httptest.NewRequest("POST", "/admin/imagebbs/approve/4", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"post": "4"})

	if result := approvePostTask.Action(httptest.NewRecorder(), req); result != nil {
		t.Fatalf("unexpected result %v", result)
	}
	if len(stub.AdminApproveImagePostCalls) != 1 || stub.AdminApproveImagePostCalls[0] != 4 {
		t.Fatalf("unexpected approve calls: %#v", stub.AdminApproveImagePostCalls)
	}
}
