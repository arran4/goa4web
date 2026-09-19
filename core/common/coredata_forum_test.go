package common

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

type mockQuerierSub struct {
	db.QuerierStub
	ListSubscribersForPatternFn func(ctx context.Context, arg db.ListSubscribersForPatternParams) ([]int32, error)
	InsertSubscriptionParams    []db.InsertSubscriptionParams
}

func (m *mockQuerierSub) ListSubscribersForPattern(ctx context.Context, arg db.ListSubscribersForPatternParams) ([]int32, error) {
	if m.ListSubscribersForPatternFn != nil {
		return m.ListSubscribersForPatternFn(ctx, arg)
	}
	return nil, sql.ErrNoRows
}

func (m *mockQuerierSub) InsertSubscription(ctx context.Context, arg db.InsertSubscriptionParams) error {
	m.InsertSubscriptionParams = append(m.InsertSubscriptionParams, arg)
	return nil
}

func TestEnsureSubscriptionIdempotent(t *testing.T) {
	ctx := context.Background()

	t.Run("AlreadySubscribed", func(t *testing.T) {
		qs := &mockQuerierSub{}
		qs.ListSubscribersForPatternFn = func(ctx context.Context, arg db.ListSubscribersForPatternParams) ([]int32, error) {
			return []int32{42}, nil
		}

		err := EnsureSubscriptionIdempotent(ctx, qs, 42, "test", "internal")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		if len(qs.InsertSubscriptionParams) != 0 {
			t.Fatalf("expected 0 inserts, got %d", len(qs.InsertSubscriptionParams))
		}
	})

	t.Run("LookupError", func(t *testing.T) {
		qs := &mockQuerierSub{}
		expectedErr := errors.New("db offline")
		qs.ListSubscribersForPatternFn = func(ctx context.Context, arg db.ListSubscribersForPatternParams) ([]int32, error) {
			return nil, expectedErr
		}

		err := EnsureSubscriptionIdempotent(ctx, qs, 42, "test", "internal")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
		if len(qs.InsertSubscriptionParams) != 0 {
			t.Fatalf("expected 0 inserts, got %d", len(qs.InsertSubscriptionParams))
		}
	})

	t.Run("InsertSuccess", func(t *testing.T) {
		qs := &mockQuerierSub{}
		qs.ListSubscribersForPatternFn = func(ctx context.Context, arg db.ListSubscribersForPatternParams) ([]int32, error) {
			return nil, sql.ErrNoRows
		}

		err := EnsureSubscriptionIdempotent(ctx, qs, 42, "test", "internal")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		if len(qs.InsertSubscriptionParams) != 1 {
			t.Fatalf("expected 1 insert, got %d", len(qs.InsertSubscriptionParams))
		}
	})
}
