package dbstart

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/handlers"
)

func TestEnsureSchemaVersionMatch(t *testing.T) {
	fq := &fakeQuerier{versionValue: handlers.ExpectedSchemaVersion}

	if err := ensureSchemaWithQuerier(context.Background(), fq); err != nil {
		t.Fatalf("ensureSchema: %v", err)
	}
}

func TestEnsureSchemaVersionMismatch(t *testing.T) {
	fq := &fakeQuerier{versionValue: handlers.ExpectedSchemaVersion - 1}

	err := ensureSchemaWithQuerier(context.Background(), fq)
	if err == nil {
		t.Fatalf("expected error")
	}
	expected := RenderSchemaMismatch(handlers.ExpectedSchemaVersion-1, handlers.ExpectedSchemaVersion)
	if err.Error() != expected {
		t.Fatalf("unexpected error: %v", err)
	}
}
