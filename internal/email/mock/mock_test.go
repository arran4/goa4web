package mock

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/email"
)

func TestProvider(t *testing.T) {
	p := &Provider{}
	fromAddr := email.ParseAddress("from@test")
	if fromAddr.Name == "" {
		fromAddr.Name = email.DefaultFromName
	}
	toAddr := email.ParseAddress("to")
	raw, err := email.BuildMessage(fromAddr, toAddr, "sub", "body", "html")
	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	if err := p.Send(context.Background(), "to", "sub", raw); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(p.Messages) != 1 {
		t.Fatalf("messages len=%d", len(p.Messages))
	}
	msg := p.Messages[0]
	if msg.To != "to" || msg.Subject != "sub" || msg.Text != "body" || msg.HTML != "html" {
		t.Fatalf("unexpected message: %#v", msg)
	}
}
