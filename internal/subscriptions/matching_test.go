package subscriptions

import (
	"testing"
)

func TestMatchDefinition_Repro(t *testing.T) {
	tests := []struct {
		pattern string
		shouldMatch bool
		definitionName string
	}{
		{
			pattern: "create thread:/forum/topic/123",
			shouldMatch: true,
			definitionName: "New Threads (Specific Topic)",
		},
		{
			pattern: "create thread:/forum/topic/123/some/other",
			shouldMatch: true,
			definitionName: "New Threads (Specific Topic)",
		},
		{
			pattern: "create thread:/forum/topic/*",
			shouldMatch: true,
			definitionName: "New Threads (All)",
		},
		{
			pattern: "reply:/forum/topic/123/thread/456",
			shouldMatch: true,
			definitionName: "Replies (Specific Thread)",
		},
		{
			pattern: "reply:/forum/topic/123/thread/456/something",
			shouldMatch: true,
			definitionName: "Replies (Specific Thread)",
		},
		{
			pattern: "reply:/forum/topic/*/thread/*",
			shouldMatch: true,
			definitionName: "Replies (All)",
		},
		{
			pattern: "edit reply:/forum/topic/123/thread/456",
			shouldMatch: true,
			definitionName: "Edit Reply",
		},
		{
			pattern: "create thread:/private/topic/123",
			shouldMatch: true,
			definitionName: "New Threads (Private Topic)",
		},
		{
			pattern: "reply:/private/topic/123/thread/456",
			shouldMatch: true,
			definitionName: "Replies (Private Thread)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			def, params := MatchDefinition(tt.pattern)
			if tt.shouldMatch {
				if def == nil {
					t.Errorf("Expected pattern %q to match a definition, but it matched nothing", tt.pattern)
				} else if def.Name != tt.definitionName {
					t.Errorf("Expected pattern %q to match definition %q, but matched %q", tt.pattern, tt.definitionName, def.Name)
				}
				t.Logf("Pattern %q matched %q with params %v", tt.pattern, def.Name, params)
			} else {
				if def != nil {
					t.Errorf("Expected pattern %q NOT to match any definition, but matched %q", tt.pattern, def.Name)
				}
			}
		})
	}
}
