package scenario

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestComposeValidatesOrderedCrossScenarioReferencesAndAssets(t *testing.T) {
	base, err := Parse([]byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: base

-- 010-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: test-password
At: 2026-01-01T00:00:00Z

`), nil)
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	feature, err := Parse([]byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: feature

-- 010-bob.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: test-password
At: 2026-01-01T00:02:00Z

-- 020-post.event --
Op: private-forum.create
Ref: room
Actor: alice
Participant: bob
Title: Room
At: 2026-01-01T00:03:00Z

-- 030-post.event --
Op: forum.post
Actor: alice
Forum: room
At: 2026-01-01T00:04:00Z
Attachment: assets/feature.jpg
`), fstest.MapFS{"assets/feature.jpg": &fstest.MapFile{Data: []byte("image")}})
	if err != nil {
		t.Fatalf("parse feature: %v", err)
	}

	combined, err := Compose(base, feature)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	if err := Validate(combined); err != nil {
		t.Fatalf("validate composed scenarios: %v", err)
	}
	if got, want := combined.Meta.Name, "base + feature"; got != want {
		t.Errorf("name = %q, want %q", got, want)
	}
	if got, want := combined.Events[1].File, "feature/010-bob.event"; got != want {
		t.Errorf("event file = %q, want %q", got, want)
	}
}

func TestComposeRejectsDuplicateReferencesAcrossScenarios(t *testing.T) {
	first, err := Parse([]byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: first

-- 010-user.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: test-password
At: 2026-01-01T00:00:00Z
`), nil)
	if err != nil {
		t.Fatalf("parse first: %v", err)
	}
	second, err := Parse([]byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: second

-- 010-user.event --
Op: user.create
Ref: alice
Username: alice2
Email: alice2@example.test
Password: test-password
At: 2026-01-01T00:01:00Z
`), nil)
	if err != nil {
		t.Fatalf("parse second: %v", err)
	}
	combined, err := Compose(first, second)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	var duplicate ErrDuplicateRef
	if err := Validate(combined); !errors.As(err, &duplicate) {
		t.Fatalf("validate error = %v, want ErrDuplicateRef", err)
	}
}
