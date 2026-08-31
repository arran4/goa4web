package scenario

import (
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestParseValidScenario(t *testing.T) {
	txt := `# Test scenario
-- scenario.meta --
Format: goa4web-scenario/v1
Name: test-scenario
Description: A test scenario description

-- 010-user.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.com
At: 2026-08-01T09:00:00Z

-- 020-post.event --
Op: forum.post
Ref: post1
Actor: alice
Forum: general
At: 2026-08-01T09:05:00Z
Attachment: assets/welcome.jpg
Attachment: assets/map.jpg

Line 1 of post body.
Line 2 of post body.
`

	fsys := fstest.MapFS{
		"assets/welcome.jpg": &fstest.MapFile{Data: []byte("jpg1")},
		"assets/map.jpg":     &fstest.MapFile{Data: []byte("jpg2")},
	}

	sc, err := Parse([]byte(txt), fsys)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if sc.Meta.Format != FormatV1 {
		t.Errorf("expected format %s, got %s", FormatV1, sc.Meta.Format)
	}
	if sc.Meta.Name != "test-scenario" {
		t.Errorf("expected name %s, got %s", "test-scenario", sc.Meta.Name)
	}
	if sc.Meta.Description != "A test scenario description" {
		t.Errorf("expected description %s, got %s", "A test scenario description", sc.Meta.Description)
	}

	// Test event ordering
	if len(sc.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(sc.Events))
	}
	if sc.Events[0].File != "010-user.event" || sc.Events[0].Op != "user.create" {
		t.Errorf("event[0] mismatch: %+v", sc.Events[0])
	}
	if sc.Events[1].File != "020-post.event" || sc.Events[1].Op != "forum.post" {
		t.Errorf("event[1] mismatch: %+v", sc.Events[1])
	}

	// Test timestamp parsing
	expectedTime := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	if !sc.Events[0].At.Equal(expectedTime) {
		t.Errorf("expected time %v, got %v", expectedTime, sc.Events[0].At)
	}

	// Test repeated header handling
	attachments := sc.Events[1].Headers.Values("Attachment")
	if len(attachments) != 2 || attachments[0] != "assets/welcome.jpg" || attachments[1] != "assets/map.jpg" {
		t.Errorf("unexpected attachments: %v", attachments)
	}

	// Test multiline body preservation
	expectedBody := "Line 1 of post body.\nLine 2 of post body."
	if sc.Events[1].Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, sc.Events[1].Body)
	}
}

func TestParseFS(t *testing.T) {
	fsys := fstest.MapFS{
		"scenarios/test/scenario.txtar": &fstest.MapFile{
			Data: []byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: fs-test

-- 01-user.event --
Op: user.create
Ref: bob
Username: bob
At: 2026-08-01T10:00:00Z
`),
		},
		"scenarios/test/assets/test.jpg": &fstest.MapFile{Data: []byte("asset")},
	}

	sc, err := ParseFS(fsys, "scenarios/test/scenario.txtar")
	if err != nil {
		t.Fatalf("ParseFS failed: %v", err)
	}
	if sc.Meta.Name != "fs-test" {
		t.Errorf("expected name fs-test, got %s", sc.Meta.Name)
	}
	if len(sc.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(sc.Events))
	}
}

func TestParseMissingMeta(t *testing.T) {
	txt := `-- 01-user.event --
Op: user.create
Username: alice
At: 2026-08-01T09:00:00Z
`
	_, err := Parse([]byte(txt), nil)
	if err == nil || !strings.Contains(err.Error(), "scenario.meta") {
		t.Fatalf("expected missing meta error, got: %v", err)
	}
}

func TestParseInvalidFormat(t *testing.T) {
	txt := `-- scenario.meta --
Format: goa4web-scenario/v999
Name: invalid

-- 01-user.event --
Op: user.create
Username: alice
At: 2026-08-01T09:00:00Z
`
	_, err := Parse([]byte(txt), nil)
	if err == nil || !strings.Contains(err.Error(), "invalid scenario format") {
		t.Fatalf("expected invalid format error, got: %v", err)
	}
}

func TestParseMalformedHeader(t *testing.T) {
	txt := `-- scenario.meta --
Format: goa4web-scenario/v1
Name: test

-- 01-user.event --
Op user.create
`
	_, err := Parse([]byte(txt), nil)
	if err == nil || !strings.Contains(err.Error(), "malformed header") {
		t.Fatalf("expected malformed header error, got: %v", err)
	}
}
