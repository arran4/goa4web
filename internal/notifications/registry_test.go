package notifications

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"
)

func TestParseTxtarConfig(t *testing.T) {
	data := []byte(`EventPattern: thread_replied
DefaultRoles: user, admin
RequiredTiers: premium
-- email.gotxt --
Subject: Hello
-- internal.gohtml --
<b>Hello</b>
`)

	config, err := ParseTxtarConfig(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.EventPattern != "thread_replied" {
		t.Errorf("expected thread_replied, got %s", config.EventPattern)
	}

	if len(config.DefaultRoles) != 2 || config.DefaultRoles[0] != "user" || config.DefaultRoles[1] != "admin" {
		t.Errorf("unexpected DefaultRoles: %v", config.DefaultRoles)
	}

	if len(config.RequiredTiers) != 1 || config.RequiredTiers[0] != "premium" {
		t.Errorf("unexpected RequiredTiers: %v", config.RequiredTiers)
	}

	if config.Templates["email.gotxt"] != "Subject: Hello" {
		t.Errorf("expected Subject: Hello, got %s", config.Templates["email.gotxt"])
	}

	if config.Templates["internal.gohtml"] != "<b>Hello</b>" {
		t.Errorf("expected <b>Hello</b>, got %s", config.Templates["internal.gohtml"])
	}
}

func TestMemoryRegistry_ProcessEvent(t *testing.T) {
	r := NewRegistry(nil)
	err := r.ProcessEvent(context.Background(), eventbus.TaskEvent{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
