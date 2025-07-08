package mock

import (
	"context"
	"net/mail"
	"testing"
)

func TestProvider(t *testing.T) {
	p := &Provider{}
	if err := p.Send(context.Background(), mail.Address{Address: "to"}, []byte("raw")); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(p.Messages) != 1 {
		t.Fatalf("messages len=%d", len(p.Messages))
	}
	msg := p.Messages[0]
	if msg.To.Address != "to" || string(msg.Raw) != "raw" {
		t.Fatalf("unexpected message: %#v", msg)
	}
}
