package main

import (
	"context"
	"testing"
)

type stubEnqueuer struct{ email, page string }

func (s *stubEnqueuer) EnqueueEmail(ctx context.Context, email, page string) error {
	s.email = email
	s.page = page
	return nil
}

func TestNotifyChange(t *testing.T) {
	st := &stubEnqueuer{}
	ctx := context.WithValue(context.Background(), ContextValues("queries"), st)
	err := notifyChange(ctx, logMailProvider{}, "a@b.com", "http://host")
	if err != nil {
		t.Fatalf("notify error: %v", err)
	}
	if st.email != "a@b.com" || st.page != "http://host" {
		t.Fatalf("record %+v", st)
	}
}

func TestNotifyChangeErrors(t *testing.T) {
	st := &stubEnqueuer{}
	ctx := context.WithValue(context.Background(), ContextValues("queries"), st)
	if err := notifyChange(ctx, logMailProvider{}, "", "p"); err == nil {
		t.Fatal("expected error for empty email")
	}
	if err := notifyChange(ctx, nil, "a@b", "p"); err == nil {
		t.Fatal("expected error for nil provider")
	}
}
