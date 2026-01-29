package common

import (
	"context"
	"testing"
    "github.com/arran4/goa4web/config"
)

func TestAbsoluteURL(t *testing.T) {
	cd := &CoreData{
        ctx: context.Background(),
        Config: &config.RuntimeConfig{
            HTTPHostname: "http://example.com",
        },
    }
    // Mock lazy value for absoluteURLBase
    cd.absoluteURLBase.Set("http://example.com")

	tests := []struct {
		name     string
		ops      []any
		expected string
	}{
		{
			name:     "Simple path",
			ops:      []any{"/foo/bar"},
			expected: "http://example.com/foo/bar",
		},
		{
			name:     "Path with anchor",
			ops:      []any{"/foo/bar#baz"},
			expected: "http://example.com/foo/bar#baz",
		},
		{
			name:     "Path with query",
			ops:      []any{"/foo/bar?q=1"},
			expected: "http://example.com/foo/bar?q=1",
		},
		{
			name:     "Path with anchor and query",
			ops:      []any{"/foo/bar?q=1#baz"},
			expected: "http://example.com/foo/bar?q=1#baz",
		},
        {
            name: "Forum reply style",
            ops: []any{"/forum/topic/1/thread/1#c9"},
            expected: "http://example.com/forum/topic/1/thread/1#c9",
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cd.AbsoluteURL(tt.ops...)
			if got != tt.expected {
				t.Errorf("AbsoluteURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}
