package mock

import (
	"bytes"
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/email"
)

func TestProvider(t *testing.T) {
	p := &Provider{}
	msg, err := email.BuildMessage("from@test", "to", "sub", "body", "html")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := p.Send(context.Background(), "to", "sub", msg); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(p.Messages) != 1 {
		t.Fatalf("messages len=%d", len(p.Messages))
	}
	rec := p.Messages[0]
	if rec.To != "to" || rec.Subject != "sub" {
		t.Fatalf("unexpected message: %#v", rec)
	}
	if !bytes.Contains(rec.Raw, []byte("body")) || !bytes.Contains(rec.Raw, []byte("html")) {
		t.Fatalf("raw body not found: %s", string(rec.Raw))
	}
}
