package scenario

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

func TestValidateExampleScenario(t *testing.T) {
	txt := `# Human-readable description.

-- scenario.meta --
Format: goa4web-scenario/v1
Name: private-forum

-- 010-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-test
At: 2026-08-01T09:00:00+10:00

-- 020-enable-alice.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:02:00+10:00

-- 030-bob.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: bob-test
At: 2026-08-01T09:03:00+10:00

-- 040-enable-bob.event --
Op: user.enable
Actor: admin
User: bob
At: 2026-08-01T09:04:00+10:00

-- 050-private-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Participant: bob
Title: Staff Room
Description: Staff forum
At: 2026-08-01T09:05:00+10:00

-- 060-welcome-post.event --
Op: forum.post
Ref: welcome
Actor: alice
Forum: staff-room
At: 2026-08-01T09:10:00+10:00
Attachment: assets/welcome.jpg

Welcome to the staff forum.
`

	fsys := fstest.MapFS{
		"assets/welcome.jpg": &fstest.MapFile{Data: []byte("jpg")},
	}

	sc, err := Parse([]byte(txt), fsys)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := Validate(sc); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestValidateErrors(t *testing.T) {
	fsys := fstest.MapFS{
		"assets/welcome.jpg": &fstest.MapFile{Data: []byte("jpg")},
	}

	tests := []struct {
		name      string
		txtar     string
		errSubstr string
	}{
		{
			name: "unknown Op rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: magic.dance
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "unknown operation \"magic.dance\"",
		},
		{
			name: "missing required Username rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "missing required field \"Username\"",
		},
		{
			name: "missing required Email in user.create rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Username: alice
Password: secret-password
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "missing required field \"Email\"",
		},
		{
			name: "missing required Password in user.create rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Username: alice
Email: alice@example.test
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "missing required field \"Password\"",
		},
		{
			name: "blank Password in user.create rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Username: alice
Email: alice@example.test
Password:   
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "missing required field \"Password\"",
		},
		{
			name: "unknown field rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Username: alice
Email: alice@example.test
Password: secret-password
SuperPower: flying
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "unknown field \"SuperPower\"",
		},
		{
			name: "duplicate Ref rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass-one
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.create
Ref: alice
Username: alice2
Email: alice2@example.test
Password: pass-two
At: 2026-08-01T09:05:00Z
`,
			errSubstr: "duplicate user reference \"alice\"",
		},
		{
			name: "unresolved reference rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.enable
User: ghost
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "unresolved user reference \"ghost\"",
		},
		{
			name: "forward reference rejected (order matters)",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.enable
User: alice
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: secret-pass
At: 2026-08-01T09:05:00Z
`,
			errSubstr: "unresolved user reference \"alice\"",
		},
		{
			name: "wrong reference type rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: secret-pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: forum.post
Ref: post1
Actor: alice
Forum: alice
At: 2026-08-01T09:05:00Z
`,
			errSubstr: "is a user, expected forum",
		},
		{
			name: "missing At timestamp rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Username: alice
Email: alice@example.test
Password: secret-pass
`,
			errSubstr: "missing required field \"At\"",
		},
		{
			name: "private-forum.create without Participant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Title: Forum
At: 2026-08-01T09:05:00Z
`,
			errSubstr: "missing required field \"Participant\"",
		},
		{
			name: "private-forum.create with Actor as Participant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Participant: alice
Title: Forum
At: 2026-08-01T09:05:00Z
`,
			errSubstr: "cannot specify Actor \"alice\" as Participant",
		},
		{
			name: "private-forum.create with duplicate Participant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: pass
At: 2026-08-01T09:01:00Z

-- 03.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Participant: bob
Participant: bob
Title: Forum
At: 2026-08-01T09:05:00Z
`,
			errSubstr: "duplicate participant \"bob\"",
		},
		{
			name: "private-forum.create with unresolved Participant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Participant: charlie
Title: Forum
At: 2026-08-01T09:05:00Z
`,
			errSubstr: "unresolved user reference \"charlie\" in field \"Participant\"",
		},
		{
			name: "private-forum.create with unresolved Actor rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: private-forum.create
Ref: forum1
Actor: charlie
Participant: bob
Title: Forum
At: 2026-08-01T09:05:00Z
`,
			errSubstr: "unresolved user reference \"charlie\" in field \"Actor\"",
		},
		{
			name: "missing asset rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: pass
At: 2026-08-01T09:01:00Z

-- 03.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Participant: bob
Title: Forum
At: 2026-08-01T09:05:00Z

-- 04.event --
Op: forum.post
Actor: alice
Forum: forum1
Attachment: assets/missing.png
At: 2026-08-01T09:10:00Z
`,
			errSubstr: "asset not found: assets/missing.png",
		},
		{
			name: "asset path escape rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: pass
At: 2026-08-01T09:01:00Z

-- 03.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Participant: bob
Title: Forum
At: 2026-08-01T09:05:00Z

-- 04.event --
Op: forum.post
Actor: alice
Forum: forum1
Attachment: ../../secret.jpg
At: 2026-08-01T09:10:00Z
`,
			errSubstr: "escapes scenario directory",
		},
		{
			name: "missing required User in user.grant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.grant
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "missing required field \"User\"",
		},
		{
			name: "missing required Section in user.grant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.grant
User: alice
Action: see
At: 2026-08-01T09:01:00Z
`,
			errSubstr: "missing required field \"Section\"",
		},
		{
			name: "missing required Action in user.grant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.grant
User: alice
Section: privateforum
At: 2026-08-01T09:01:00Z
`,
			errSubstr: "missing required field \"Action\"",
		},
		{
			name: "unknown field in user.grant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.grant
User: alice
Section: privateforum
Action: see
UnknownKey: foo
At: 2026-08-01T09:01:00Z
`,
			errSubstr: "unknown field \"UnknownKey\"",
		},
		{
			name: "unresolved user ref in user.grant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.grant
User: unknown-user
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "unresolved user reference \"unknown-user\" in field \"User\"",
		},
		{
			name: "invalid permission tuple in user.grant rejected",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.grant
User: alice
Section: invalid_section
Item: topic
Action: see
At: 2026-08-01T09:01:00Z
`,
			errSubstr: "invalid or unsupported permission tuple",
		},
		{
			name: "item-specific privateforum_thread view permission rejected in user.grant",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.grant
User: alice
Section: privateforum_thread
Item: thread
Action: view
At: 2026-08-01T09:01:00Z
`,
			errSubstr: "requires a concrete item ID and cannot be granted globally",
		},
		{
			name: "item-specific forum topic post permission rejected in user.grant",
			txtar: `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.grant
User: alice
Section: forum
Item: topic
Action: post
At: 2026-08-01T09:01:00Z
`,
			errSubstr: "requires a concrete item ID and cannot be granted globally",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc, err := Parse([]byte(tt.txtar), fsys)
			if err != nil {
				// Parse error might happen (e.g. malformed headers)
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("Parse error %v, want substring %q", err, tt.errSubstr)
				}
				return
			}
			err = Validate(sc)
			if err == nil {
				t.Fatalf("Validate expected error containing %q, got nil", tt.errSubstr)
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Fatalf("Validate error %v, want substring %q", err, tt.errSubstr)
			}
		})
	}
}

func TestValidatePrivateForumMultipleParticipants(t *testing.T) {
	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test-multiple-participants

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass-alice
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: pass-bob
At: 2026-08-01T09:01:00Z

-- 03.event --
Op: user.create
Ref: charlie
Username: charlie
Email: charlie@example.test
Password: pass-charlie
At: 2026-08-01T09:02:00Z

-- 04.event --
Op: private-forum.create
Ref: group-chat
Actor: alice
Participant: bob
Participant: charlie
Title: Group Chat
At: 2026-08-01T09:05:00Z
`

	sc, err := Parse([]byte(txt), nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if err := Validate(sc); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	data, ok := sc.Events[3].OpData.(*PrivateForumCreateData)
	if !ok {
		t.Fatalf("expected PrivateForumCreateData, got: %T", sc.Events[3].OpData)
	}
	if len(data.Participants) != 2 || data.Participants[0] != "bob" || data.Participants[1] != "charlie" {
		t.Errorf("unexpected participants: %v", data.Participants)
	}
}

func TestValidateNilScenario(t *testing.T) {
	if err := Validate(nil); err == nil {
		t.Fatal("expected error for nil scenario, got nil")
	}
}

func TestValidateInvalidMeta(t *testing.T) {
	s := &Scenario{
		Meta: Meta{
			Format: FormatV1,
			Name:   "",
		},
	}
	var errMissing ErrMissingRequiredField
	if err := Validate(s); !errors.As(err, &errMissing) {
		t.Fatalf("expected ErrMissingRequiredField, got %v", err)
	}
}

func TestValidateUserGrant(t *testing.T) {
	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test-user-grant

-- 01.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:01:00Z

-- 03.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: create
At: 2026-08-01T09:02:00Z
`

	sc, err := Parse([]byte(txt), nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if err := Validate(sc); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	d1, ok := sc.Events[1].OpData.(*UserGrantData)
	if !ok || d1.User != "alice" || d1.Section != "privateforum" || d1.Item != "topic" || d1.Action != "see" {
		t.Errorf("unexpected d1: %+v", d1)
	}

	d2, ok := sc.Events[2].OpData.(*UserGrantData)
	if !ok || d2.User != "alice" || d2.Section != "privateforum" || d2.Item != "topic" || d2.Action != "create" {
		t.Errorf("unexpected d2: %+v", d2)
	}
}
