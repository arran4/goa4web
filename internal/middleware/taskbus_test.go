package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

func TestTaskEventMiddleware(t *testing.T) {
	bus := eventbus.NewBus()
	eventbus.DefaultBus = bus
	defer func() { eventbus.DefaultBus = eventbus.NewBus() }()

	successHandler := TaskEventMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/admin/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ch := bus.Subscribe()
	successHandler.ServeHTTP(rec, req)
	select {
	case evt := <-ch:
		if evt.Task != "Add" || evt.Path != "/admin/p" || !evt.Admin {
			t.Fatalf("unexpected event %+v", evt)
		}
	default:
		t.Fatalf("expected event on success")
	}

	// non-admin path should not set Admin flag
	req = httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ch = bus.Subscribe()
	successHandler.ServeHTTP(rec, req)
	select {
	case evt := <-ch:
		if evt.Admin {
			t.Fatalf("unexpected admin flag for %#v", evt)
		}
	default:
		t.Fatalf("expected event for non-admin path")
	}

	failureHandler := TaskEventMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	req = httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ch = bus.Subscribe()
	failureHandler.ServeHTTP(rec, req)
	select {
	case <-ch:
		t.Fatalf("did not expect event on failure")
	default:
	}

	// ensure handlers can attach event data
	itemHandler := TaskEventMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if evt, ok := r.Context().Value(hcommon.KeyBusEvent).(*eventbus.Event); ok {
			evt.Item = "info"
		}
		w.WriteHeader(http.StatusOK)
	}))
	req = httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ch = bus.Subscribe()
	itemHandler.ServeHTTP(rec, req)
	select {
	case evt := <-ch:
		if evt.Item != "info" {
			t.Fatalf("missing item: %+v", evt)
		}
	default:
		t.Fatalf("expected event with item")
	}
}
