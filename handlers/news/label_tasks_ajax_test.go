package news

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/mux"
)

func TestMarkReadTaskAjax(t *testing.T) {
	cd := &common.CoreData{}

	task := &MarkReadTask{}

	req := httptest.NewRequest("GET", "/news/123/labels?task=Mark+Thread+Read&ajax=1", nil)

	// Add context
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	// Add vars using gorilla mux
	vars := map[string]string{
		"news": "123",
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	res := task.Action(w, req)

	handlerFunc, ok := res.(http.HandlerFunc)
	if !ok {
		t.Fatalf("Expected http.HandlerFunc, got %T", res)
	}

	handlerFunc(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	if w.Body.String() != "" {
		t.Errorf("Expected empty body, got %q", w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html; charset=utf-8, got %q", contentType)
	}
}
