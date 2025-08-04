package common

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
)

func TestWithSelectionsFromRequest(t *testing.T) {
	cfg := config.NewRuntimeConfig()

	t.Run("path variable", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req = mux.SetURLVars(req, map[string]string{"board": "1"})
		cd := NewCoreData(context.Background(), nil, cfg, WithSelectionsFromRequest(req))
		if got := cd.currentBoardID; got != 1 {
			t.Fatalf("currentBoardID = %d; want 1", got)
		}
	})

	t.Run("query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?thread=2", nil)
		cd := NewCoreData(context.Background(), nil, cfg, WithSelectionsFromRequest(req))
		if got := cd.currentThreadID; got != 2 {
			t.Fatalf("currentThreadID = %d; want 2", got)
		}
	})

	t.Run("form value", func(t *testing.T) {
		body := strings.NewReader("post=3")
		req := httptest.NewRequest("POST", "/", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := NewCoreData(context.Background(), nil, cfg, WithSelectionsFromRequest(req))
		if got := cd.currentImagePostID; got != 3 {
			t.Fatalf("currentImagePostID = %d; want 3", got)
		}
	})
}
