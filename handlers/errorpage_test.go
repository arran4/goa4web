package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/handlertest"
)

func TestRenderErrorPageNotFoundOmitsInternalError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	req, cd, _ := handlertest.RequestWithCoreData(t, req)

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
