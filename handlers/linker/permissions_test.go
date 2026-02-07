package linker

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUserCanCreateLink(t *testing.T) {
	t.Run("Allowed", userCanCreateLinkAllowed)
	t.Run("Denied", userCanCreateLinkDenied)
	t.Run("Error", userCanCreateLinkError)
}

func userCanCreateLinkAllowed(t *testing.T) {
	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrantResult(true),
	)

	ok, err := UserCanCreateLink(context.Background(), q, sql.NullInt32{Int32: 1, Valid: true}, 2)
	if err != nil {
		t.Fatalf("UserCanCreateLink: %v", err)
	}
	if !ok {
		t.Errorf("expected allowed")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func userCanCreateLinkDenied(t *testing.T) {
	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrantError(sql.ErrNoRows),
	)

	ok, err := UserCanCreateLink(context.Background(), q, sql.NullInt32{Int32: 1, Valid: true}, 2)
	if err != nil {
		t.Fatalf("UserCanCreateLink: %v", err)
	}
	if ok {
		t.Errorf("expected denied")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func userCanCreateLinkError(t *testing.T) {
	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrantError(errors.New("db offline")),
	)

	ok, err := UserCanCreateLink(context.Background(), q, sql.NullInt32{Int32: 1, Valid: true}, 2)
	if err == nil {
		t.Fatalf("expected error")
	}
	if ok {
		t.Fatalf("expected denied when error occurs")
	}
}
