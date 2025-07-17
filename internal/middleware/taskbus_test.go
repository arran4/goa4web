package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		if cd, ok := r.Context().Value(corecorecommon.KeyCoreData).(*corecorecommon.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["info"] = true
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	req = httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ch = bus.Subscribe()
	ctx := context.WithValue(req.Context(), corecorecommon.KeyCoreData, &corecorecommon.CoreData{})
	itemHandler.ServeHTTP(rec, req.WithContext(ctx))
	select {
	case evt := <-ch:
		val, ok := evt.Data["info"].(bool)
		if evt.Data == nil || !ok || !val {
			t.Fatalf("missing data: %+v", evt)
		}
	default:
		t.Fatalf("expected event with data")
	}
}

func TestTaskEventQueue(t *testing.T) {
	bus := eventbus.NewBus()
	eventbus.DefaultBus = bus
	defer func() { eventbus.DefaultBus = eventbus.NewBus() }()

	taskQueue = newEventQueue(maxQueuedTaskEvents)

	if err := bus.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown bus: %v", err)
	}

	handler := TaskEventMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if len(taskQueue.events) != 1 {
		t.Fatalf("expected queued event")
	}

	eventbus.ReopenDefaultBus()
	ch := eventbus.DefaultBus.Subscribe()
	taskQueue.flush(context.Background())

	select {
	case <-ch:
	default:
		t.Fatalf("expected flushed event")
	}
}
