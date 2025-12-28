package handlers

import (
	"context"
	"errors"
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

func TestRenderErrorPageNotFoundOmitsInternalError(t *testing.T) {
	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	cd := common.NewCoreData(req.Context(), db.New(conn), config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	data := struct {
		*common.CoreData
		Error   string
		BackURL string
	}{
		CoreData: cd,
		Error:    "",
		BackURL:  "",
	}
	rr := httptest.NewRecorder()
	if err := cd.ExecuteSiteTemplate(rr, req, "notFoundPage.gohtml", data); err != nil {
		t.Fatalf("ExecuteSiteTemplate: %v", err)
	}
	rr = httptest.NewRecorder()
	RenderErrorPage(rr, req, WrapNotFound(errors.New("Internal Server Error")))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d got %d", http.StatusNotFound, rr.Code)
	}
	if strings.Contains(rr.Body.String(), "Internal Server Error") {
		t.Fatalf("expected 404 page to omit internal error message")
	}
}
