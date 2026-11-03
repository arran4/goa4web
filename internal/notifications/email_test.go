package notifications

import (
	"bytes"
	"context"
	"mime"
	"net/mail"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/testhelpers"
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

func TestRenderEmailFromTemplatesAutomaticMetadata(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.EmailFrom = "from@example.com"
	n := New(WithConfig(cfg))

	manual, err := n.RenderEmailFromTemplates(context.Background(), "to@example.com", &EmailTemplates{}, nil)
	if err != nil {
		t.Fatalf("render manual email: %v", err)
	}
	manualMessage, err := mail.ReadMessage(bytes.NewReader(manual))
	if err != nil {
		t.Fatalf("read manual email: %v", err)
	}
	if got := manualMessage.Header.Get("Auto-Submitted"); got != "" {
		t.Errorf("manual Auto-Submitted=%q, want empty", got)
	}

	automatic, err := n.RenderEmailFromTemplates(context.Background(), "to@example.com", &EmailTemplates{}, nil, WithAutomaticDelivery())
	if err != nil {
		t.Fatalf("render automatic email: %v", err)
	}
	automaticMessage, err := mail.ReadMessage(bytes.NewReader(automatic))
	if err != nil {
		t.Fatalf("read automatic email: %v", err)
	}
	if got := automaticMessage.Header.Get("Auto-Submitted"); got != "auto-generated" {
		t.Errorf("automatic Auto-Submitted=%q, want auto-generated", got)
	}
}

func TestRenderAndQueueEmailFromTemplatesMarksNotificationAutomatic(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.EmailFrom = "from@example.com"
	queries := testhelpers.NewQuerierStub()
	n := New(WithConfig(cfg), WithQueries(queries))
	userID := int32(1)

	if err := n.renderAndQueueEmailFromTemplates(context.Background(), &userID, "to@example.com", &EmailTemplates{}, nil); err != nil {
		t.Fatalf("renderAndQueueEmailFromTemplates: %v", err)
	}
	if len(queries.InsertPendingEmailCalls) != 1 {
		t.Fatalf("queued emails=%d, want 1", len(queries.InsertPendingEmailCalls))
	}
	queued, err := mail.ReadMessage(bytes.NewBufferString(queries.InsertPendingEmailCalls[0].Body))
	if err != nil {
		t.Fatalf("read queued email: %v", err)
	}
	if got := queued.Header.Get("Auto-Submitted"); got != "auto-generated" {
		t.Errorf("queued Auto-Submitted=%q, want auto-generated", got)
	}
}
