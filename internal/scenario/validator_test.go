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
At: 2026-08-01T09:00:00+10:00

-- 020-enable-alice.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:02:00+10:00

-- 030-private-forum.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Title: Staff Room
At: 2026-08-01T09:05:00+10:00

-- 040-welcome-post.event --
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
			name: "missing required field rejected",
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
At: 2026-08-01T09:00:00Z
`,
			errSubstr: "missing required field \"Email\"",
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
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: user.create
Ref: alice
Username: alice2
Email: alice2@example.test
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
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: forum.post
Ref: post1
Actor: admin
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
`,
			errSubstr: "missing required field \"At\"",
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
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Title: Forum
At: 2026-08-01T09:05:00Z

-- 03.event --
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
At: 2026-08-01T09:00:00Z

-- 02.event --
Op: private-forum.create
Ref: forum1
Actor: alice
Title: Forum
At: 2026-08-01T09:05:00Z

-- 03.event --
Op: forum.post
Actor: alice
Forum: forum1
Attachment: ../../secret.jpg
At: 2026-08-01T09:10:00Z
`,
			errSubstr: "escapes scenario directory",
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
