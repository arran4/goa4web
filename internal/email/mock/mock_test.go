package mock

import (
	"context"
	"testing"
)

func TestProvider(t *testing.T) {
	p := &Provider{}
	if err := p.Send(context.Background(), "to", "sub", []byte("raw")); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(p.Messages) != 1 {
		t.Fatalf("messages len=%d", len(p.Messages))
	}
	msg := p.Messages[0]
	if msg.To != "to" || msg.Subject != "sub" || string(msg.Raw) != "raw" {
		t.Fatalf("unexpected message: %#v", msg)
	}
}
