package common_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestWithSelectionsFromRequest(t *testing.T) {
	cfg := config.NewRuntimeConfig()

	t.Run("path variable", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"board": "1"})
		cd := common.NewCoreData(context.Background(), nil, cfg, common.WithSelectionsFromRequest(req))
		if got := cd.SelectedBoardID(); got != 1 {
			t.Fatalf("SelectedBoardID = %d; want 1", got)
		}
	})

	t.Run("query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?thread=2", nil)
		cd := common.NewCoreData(context.Background(), nil, cfg, common.WithSelectionsFromRequest(req))
		if got := cd.SelectedThreadID(); got != 2 {
			t.Fatalf("SelectedThreadID = %d; want 2", got)
		}
	})

	t.Run("form value", func(t *testing.T) {
		body := strings.NewReader("post=3")
		req := httptest.NewRequest("POST", "/", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(context.Background(), nil, cfg, common.WithSelectionsFromRequest(req))
		if got := cd.SelectedImagePostID(); got != 3 {
			t.Fatalf("SelectedImagePostID = %d; want 3", got)
		}
	})
}
