package externallink

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

type result struct{ id int64 }

func (r result) LastInsertId() (int64, error) { return r.id, nil }
func (result) RowsAffected() (int64, error)   { return 0, nil }

func TestResolveRecordFallsBackToExistingLink(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	qs.EnsureExternalLinkFn = func(context.Context, string) (sql.Result, error) { return result{}, nil }
	qs.GetExternalLinkFn = func(context.Context, string) (*db.ExternalLink, error) {
		return &db.ExternalLink{ID: 42}, nil
	}

	id, err := ResolveRecord(context.Background(), qs, "https://example.test")
	if err != nil || id != 42 {
		t.Fatalf("ResolveRecord() = (%d, %v), want (42, nil)", id, err)
	}
}

func TestNeedsMetadata(t *testing.T) {
	if !NeedsMetadata(nil) {
		t.Fatal("nil link should need metadata")
	}
	if !NeedsMetadata(&db.ExternalLink{CardTitle: sql.NullString{}}) {
		t.Fatal("empty title should need metadata")
	}
	if NeedsMetadata(&db.ExternalLink{CardTitle: sql.NullString{String: "title", Valid: true}}) {
		t.Fatal("populated title should not need metadata")
	}
}
