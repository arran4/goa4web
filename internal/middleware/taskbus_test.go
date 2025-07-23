package middleware

import (
	"context"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestTaskEventMiddleware(t *testing.T) {
	bus := eventbus.NewBus()

	successHandler := TaskEventMiddlewareWithBus(bus)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/admin/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ch := bus.Subscribe(eventbus.TaskMessageType)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	successHandler.ServeHTTP(rec, req.WithContext(ctx))
	select {
	case msg := <-ch:
		evt, ok := msg.(eventbus.TaskEvent)
		if !ok {
			t.Fatalf("wrong type %T", msg)
		}
		named, ok := evt.Task.(tasks.Name)
		if !ok || named.Name() != "MISSING" || evt.Path != "/admin/p" {
			t.Fatalf("unexpected event %+v", evt)
		}
	default:
		t.Fatalf("expected event on success")
	}

	// non-admin path should not include /admin prefix
	req = httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ch = bus.Subscribe(eventbus.TaskMessageType)
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	successHandler.ServeHTTP(rec, req.WithContext(ctx))
	select {
	case msg := <-ch:
		evt, ok := msg.(eventbus.TaskEvent)
		if !ok {
			t.Fatalf("wrong type %T", msg)
		}
		if strings.Contains(evt.Path, "/admin") {
			t.Fatalf("unexpected admin path for %#v", evt)
		}
	default:
		t.Fatalf("expected event for non-admin path")
	}

	failureHandler := TaskEventMiddlewareWithBus(bus)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	req = httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	ch = bus.Subscribe(eventbus.TaskMessageType)
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	failureHandler.ServeHTTP(rec, req.WithContext(ctx))
	select {
	case msg := <-ch:
		t.Fatalf("did not expect event on failure, got %T", msg)
		t.Fatalf("did not expect event on failure")
	default:
	}

	// ensure handlers can attach event data
	itemHandler := TaskEventMiddlewareWithBus(bus)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
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
	ch = bus.Subscribe(eventbus.TaskMessageType)
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	itemHandler.ServeHTTP(rec, req.WithContext(ctx))
	select {
	case msg := <-ch:
		evt, ok := msg.(eventbus.TaskEvent)
		if !ok {
			t.Fatalf("wrong type %T", msg)
		}
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

	taskQueue = newEventQueue(maxQueuedTaskEvents, bus)

	if err := bus.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown bus: %v", err)
	}

	handler := TaskEventMiddlewareWithBus(bus)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/p", strings.NewReader("task=Add"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	handler.ServeHTTP(rec, req.WithContext(ctx))

	if len(taskQueue.events) != 1 {
		t.Fatalf("expected queued event")
	}

	bus = eventbus.NewBus()
	taskQueue.bus = bus
	ch := bus.Subscribe(eventbus.TaskMessageType)
	taskQueue.flush(context.Background())

	select {
	case <-ch:
	default:
		t.Fatalf("expected flushed event")
	}
}

func TestTaskEventMiddleware_EventProvided(t *testing.T) {
	handler := TaskEventMiddlewareWithBus(eventbus.NewBus())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd == nil || cd.Event() == nil {
			t.Fatalf("missing event")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req.WithContext(ctx))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestTaskEventMiddleware_NoCoreDataPanic(t *testing.T) {
	handler := TaskEventMiddlewareWithBus(eventbus.NewBus())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not run")
	}))

	req := httptest.NewRequest("GET", "/p", nil)
	rec := httptest.NewRecorder()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	handler.ServeHTTP(rec, req)
}
