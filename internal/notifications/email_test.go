package notifications

import (
	"bytes"
	"context"
	"mime"
	"net/mail"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestRenderEmailFromTemplates_AdminSubject(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.EmailFrom = "from@example.com"
	n := New(WithConfig(cfg))
	msg, err := n.RenderEmailFromTemplates(context.Background(), "to@example.com", &EmailTemplates{}, nil, WithAdmin())
	if err != nil {
		t.Fatalf("RenderEmailFromTemplates: %v", err)
	}
	m, err := mail.ReadMessage(bytes.NewReader(msg))
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	subj, err := new(mime.WordDecoder).DecodeHeader(m.Header.Get("Subject"))
	if err != nil {
		t.Fatalf("DecodeHeader: %v", err)
	}
	want := "[goa4web Admin] Website Update Notification"
	if subj != want {
		t.Fatalf("subject=%q want %q", subj, want)
	}
}
