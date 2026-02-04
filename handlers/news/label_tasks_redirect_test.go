package news

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

func TestMarkReadTaskRedirect(t *testing.T) {
	cd := &common.CoreData{}

	task := &MarkReadTask{}

	redirectURL := "/some/where"
	req := httptest.NewRequest("GET", "/news/123/labels?task=Mark+Thread+Read&redirect="+redirectURL, nil)

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

	redirectHandler, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("Expected handlers.RefreshDirectHandler, got %T", res)
	}

	if redirectHandler.TargetURL != redirectURL {
		t.Errorf("Expected TargetURL to be %q, got %q", redirectURL, redirectHandler.TargetURL)
	}
}
