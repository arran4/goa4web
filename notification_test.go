package main

import (
	"context"
	"database/sql"
	"testing"
)

type stubSubscriber struct{}

func (stubSubscriber) ListSubscribersForThread(ctx context.Context, arg ListSubscribersForThreadParams) ([]*ListSubscribersForThreadRow, error) {
	return []*ListSubscribersForThreadRow{
		{Username: sql.NullString{String: "u1@example.com", Valid: true}},
		{Username: sql.NullString{String: "u2@example.com", Valid: true}},
	}, nil
}

func TestQueueThreadNotifications(t *testing.T) {
	clearEmailQueue()
	q := stubSubscriber{}
	queueThreadNotifications(context.Background(), q, 1, 2, "/page")
	items := getQueuedEmails()
	if len(items) != 2 {
		t.Fatalf("queued %d notifications", len(items))
	}
	if items[0].To != "u1@example.com" || items[0].URL != "/page" {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
}
