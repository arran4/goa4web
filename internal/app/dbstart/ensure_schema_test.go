package dbstart

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/handlers"
)

func TestEnsureSchemaVersionMatch(t *testing.T) {
	store := &fakeStore{version: handlers.ExpectedSchemaVersion, hasVersion: true}

	if err := ensureSchemaWithStore(context.Background(), store); err != nil {
		t.Fatalf("ensureSchema: %v", err)
	}
}

func TestEnsureSchemaVersionMismatch(t *testing.T) {
	store := &fakeStore{version: handlers.ExpectedSchemaVersion - 1, hasVersion: true}

	err := ensureSchemaWithStore(context.Background(), store)
	if err == nil {
		t.Fatalf("expected error")
	}
	expected := RenderSchemaMismatch(handlers.ExpectedSchemaVersion-1, handlers.ExpectedSchemaVersion)
	if err.Error() != expected {
		t.Fatalf("unexpected error: %v", err)
	}
}
