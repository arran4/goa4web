package writings

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUserCanCreateWriting(t *testing.T) {
	t.Run("Happy Path - Allowed", func(t *testing.T) {
		q := testhelpers.NewQuerierStub(
			testhelpers.WithGrantResult(true),
		)

		ok, err := UserCanCreateWriting(context.Background(), q, 1, 2)
		if err != nil {
			t.Fatalf("UserCanCreateWriting: %v", err)
		}
		if !ok {
			t.Errorf("expected allowed")
		}
		if len(q.SystemCheckGrantCalls) != 1 {
			t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
		}
	})

	t.Run("Unhappy Path - Denied", func(t *testing.T) {
		q := testhelpers.NewQuerierStub(
			testhelpers.WithGrantError(sql.ErrNoRows),
		)

		ok, err := UserCanCreateWriting(context.Background(), q, 1, 2)
		if err != nil {
			t.Fatalf("UserCanCreateWriting: %v", err)
		}
		if ok {
			t.Errorf("expected denied")
		}
		if len(q.SystemCheckGrantCalls) != 1 {
			t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
		}
	})

	t.Run("Unhappy Path - Error", func(t *testing.T) {
		q := testhelpers.NewQuerierStub(
			testhelpers.WithGrantError(errors.New("db offline")),
		)

		ok, err := UserCanCreateWriting(context.Background(), q, 1, 2)
		if err == nil {
			t.Fatalf("expected error")
		}
		if ok {
			t.Fatalf("expected denied when error occurs")
		}
	})
}
