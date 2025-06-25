package goa4web

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestTaskMatcher(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader("task="+TaskCreate))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("parse form: %v", err)
	}
	if !TaskMatcher(TaskCreate)(req, &mux.RouteMatch{}) {
		t.Errorf("expected task matcher to pass")
	}
	if TaskMatcher(TaskEdit)(req, &mux.RouteMatch{}) {
		t.Errorf("unexpected match")
	}
}

func TestNoTask(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("parse form: %v", err)
	}
	if !NoTask()(req, &mux.RouteMatch{}) {
		t.Errorf("expected match when no task")
	}

	req = httptest.NewRequest("POST", "/", strings.NewReader("task=x"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("parse form: %v", err)
	}
	if NoTask()(req, &mux.RouteMatch{}) {
		t.Errorf("unexpected match when task provided")
	}
}
