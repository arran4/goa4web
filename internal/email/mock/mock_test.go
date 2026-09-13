package mock

import (
	"bytes"
	"context"
	"net/mail"
	"testing"

	"github.com/arran4/goa4web/internal/email"
)

func TestProvider(t *testing.T) {
	p := &Provider{}
	from := email.ParseAddress("Tester <from@test>")
	to := email.ParseAddress("Receiver <to@test>")
	msg, err := email.BuildMessage(from, to, "sub", "body", "html")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := p.Send(context.Background(), to, msg); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(p.Messages) != 1 {
		t.Fatalf("messages len=%d", len(p.Messages))
	}
	rec := p.Messages[0]
	if rec.To != to || rec.Subject != "sub" {
		t.Fatalf("unexpected message: %#v", rec)
	}
	if rec.Text != "body" || rec.HTML != "html" {
		t.Fatalf("parsed bodies incorrect: %#v", rec)
	}

	m, err := mail.ReadMessage(bytes.NewReader(rec.Raw))
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	addr, err := mail.ParseAddress(m.Header.Get("From"))
	if err != nil || addr.Name != "Tester" {
		t.Fatalf("unexpected From header: %s", m.Header.Get("From"))
	}
	if m.Header.Get("Date") == "" {
		t.Fatalf("missing Date header")
	}
	if m.Header.Get("Message-ID") == "" {
		t.Fatalf("missing Message-ID header")
	}
}
