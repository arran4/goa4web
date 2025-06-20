package main

import (
	"context"
	"testing"
)

type recordMail struct{ to, sub, body string }

func (r *recordMail) Send(ctx context.Context, to, subject, body string) error {
	r.to, r.sub, r.body = to, subject, body
	return nil
}

func TestNotifyChange(t *testing.T) {
	rec := &recordMail{}
	err := notifyChange(context.Background(), rec, "a@b.com", "http://host")
	if err != nil {
		t.Fatalf("notify error: %v", err)
	}
	if rec.to != "a@b.com" || rec.sub == "" || rec.body == "" {
		t.Fatalf("record %+v", rec)
	}
}

func TestNotifyChangeErrors(t *testing.T) {
	rec := &recordMail{}
	if err := notifyChange(context.Background(), rec, "", "p"); err == nil {
		t.Fatal("expected error for empty email")
	}
	if err := notifyChange(context.Background(), nil, "a@b", "p"); err == nil {
		t.Fatal("expected error for nil provider")
	}
}
