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

func TestCheckBoardAccess(t *testing.T) {
	t.Run("Happy Path - Allows Access With Grant", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
			if arg.Section == "imagebbs" && arg.Item.String == "board" && arg.Action == "view" && arg.ItemID.Int32 == 1 {
				return 1, nil
			}
			return 0, sql.ErrNoRows
		}
		cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		cd.UserID = 42

		req := httptest.NewRequest("GET", "/imagebbs/board/1", nil)
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		req = mux.SetURLVars(req, map[string]string{"boardno": "1"})

		rr := httptest.NewRecorder()
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		handler := CheckBoardAccess(next)
		handler.ServeHTTP(rr, req)

		if !nextCalled {
			t.Errorf("Expected next handler to be called")
		}
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Unhappy Path - Denies Access Without Grant", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
			return 0, sql.ErrNoRows
		}
		cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		cd.UserID = 42

		req := httptest.NewRequest("GET", "/imagebbs/board/2", nil)
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		req = mux.SetURLVars(req, map[string]string{"boardno": "2"})

		rr := httptest.NewRecorder()
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		handler := CheckBoardAccess(next)
		handler.ServeHTTP(rr, req)

		if nextCalled {
			t.Errorf("Expected next handler NOT to be called")
		}
		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected status Forbidden (403), got %v", rr.Code)
		}
	})

	t.Run("Unhappy Path - 403 when ID is missing or invalid", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

		req := httptest.NewRequest("GET", "/imagebbs/board/invalid", nil)
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		req = mux.SetURLVars(req, map[string]string{"boardno": "invalid"})

		rr := httptest.NewRecorder()
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		handler := CheckBoardAccess(next)
		handler.ServeHTTP(rr, req)

		if nextCalled {
			t.Errorf("Expected next handler NOT to be called")
		}
		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected status Forbidden (403), got %v", rr.Code)
		}
	})
}

func TestImageBoardIDFromRequest(t *testing.T) {
	t.Run("Happy Path - Extracts boardno", func(t *testing.T) {
		cd := common.NewCoreData(context.Background(), testhelpers.NewQuerierStub(), config.NewRuntimeConfig())
		req := httptest.NewRequest("GET", "/board/10", nil)
		req = mux.SetURLVars(req, map[string]string{"boardno": "10"})

		id, err := imageBoardIDFromRequest(req, cd)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if id != 10 {
			t.Errorf("Expected ID 10, got %d", id)
		}
	})

	t.Run("Happy Path - Extracts board", func(t *testing.T) {
		cd := common.NewCoreData(context.Background(), testhelpers.NewQuerierStub(), config.NewRuntimeConfig())
		req := httptest.NewRequest("GET", "/board/20", nil)
		req = mux.SetURLVars(req, map[string]string{"board": "20"})

		id, err := imageBoardIDFromRequest(req, cd)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if id != 20 {
			t.Errorf("Expected ID 20, got %d", id)
		}
	})
}
