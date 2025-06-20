package main

import (
	"context"
	"testing"
)

func TestNotifyChangeQueues(t *testing.T) {
	st := &stubEnqueuer{}
	ctx := context.WithValue(context.Background(), ContextValues("queries"), st)
	err := notifyChange(ctx, logMailProvider{}, "a@b.com", "p")
	if err != nil {
		t.Fatalf("notify error: %v", err)
	}
	if st.email != "a@b.com" || st.page != "p" {
		t.Fatalf("stored %+v", st)
	}
}

type stubStore struct {
	items   []EmailQueueItem
	deleted []int64
}

func (s *stubStore) ListQueuedEmails(ctx context.Context, limit int32) ([]EmailQueueItem, error) {
	return s.items, nil
}
func (s *stubStore) DeleteQueuedEmail(ctx context.Context, id int64) error {
	s.deleted = append(s.deleted, id)
	return nil
}

type recordProvider struct{ sent int }

func (r *recordProvider) Send(ctx context.Context, to, subject, body string) error {
	r.sent++
	return nil
}

func TestProcessEmailQueue(t *testing.T) {
	store := &stubStore{items: []EmailQueueItem{{IDEmailQueue: 1, Email: "a@b", Page: "p"}}}
	prov := &recordProvider{}
	processEmailQueue(context.Background(), store, prov)
	if prov.sent != 1 {
		t.Fatalf("sent %d", prov.sent)
	}
	if len(store.deleted) != 1 || store.deleted[0] != 1 {
		t.Fatalf("deleted %+v", store.deleted)
	}
}
