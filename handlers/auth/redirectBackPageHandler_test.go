package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestRedirectBackPageHandlerGETAlt(t *testing.T) {
	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h := redirectBackPageHandler{BackURL: "/foo", Method: http.MethodGet}
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if cd.AutoRefresh == "" || !strings.Contains(cd.AutoRefresh, "url=/foo") {
		t.Fatalf("auto refresh=%q", cd.AutoRefresh)
	}
}
