package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/handlers/handlertest"
)

func TestRenderErrorPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		req, cd, _ := handlertest.RequestWithCoreData(t, req)

		rr := httptest.NewRecorder()
		RenderErrorPage(rr, req, errors.New("Some generic error"))

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d got %d", http.StatusInternalServerError, rr.Code)
		}
		if cd.PageTitle != "Error" {
			t.Fatalf("expected PageTitle %q got %q", "Error", cd.PageTitle)
		}
		if !strings.Contains(rr.Body.String(), "Some generic error") {
			t.Fatalf("expected body to contain error message")
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		req, cd, _ := handlertest.RequestWithCoreData(t, req)

		rr := httptest.NewRecorder()
		RenderErrorPage(rr, req, WrapNotFound(errors.New("Internal Server Error")))

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status %d got %d", http.StatusNotFound, rr.Code)
		}
		if cd.PageTitle != "Not Found" {
			t.Fatalf("expected PageTitle %q got %q", "Not Found", cd.PageTitle)
		}
		if strings.Contains(rr.Body.String(), "Internal Server Error") {
			t.Fatalf("expected 404 page to omit internal error message")
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req, cd, _ := handlertest.RequestWithCoreData(t, req)

		rr := httptest.NewRecorder()
		RenderErrorPage(rr, req, errors.New("Unauthorized"))

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d got %d", http.StatusUnauthorized, rr.Code)
		}
		if cd.PageTitle != "Error" {
			t.Fatalf("expected PageTitle %q got %q", "Error", cd.PageTitle)
		}
	})

	t.Run("Login Required", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req, cd, _ := handlertest.RequestWithCoreData(t, req)

		rr := httptest.NewRecorder()
		RenderErrorPage(rr, req, ErrLoginRequired)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d got %d", http.StatusInternalServerError, rr.Code)
		}
		if cd.PageTitle != "Login Required" {
			t.Fatalf("expected PageTitle %q got %q", "Login Required", cd.PageTitle)
		}
	})
}
