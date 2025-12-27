package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestQuerierStubTemplateOverrides(t *testing.T) {
	q := &QuerierStub{}

	if err := q.AdminSetTemplateOverride(context.Background(), AdminSetTemplateOverrideParams{Name: "t", Body: "body"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if got := q.TemplateOverrides["t"]; got != "body" {
		t.Fatalf("set body: %q", got)
	}

	body, err := q.SystemGetTemplateOverride(context.Background(), "t")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if body != "body" {
		t.Fatalf("get body: %q", body)
	}

	if err := q.AdminDeleteTemplateOverride(context.Background(), "t"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok := q.TemplateOverrides["t"]; ok {
		t.Fatalf("delete body still present")
	}

	_, err = q.SystemGetTemplateOverride(context.Background(), "t")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
}
