package common_test

import (
	"net/http"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestSanitizeBackURL_Vulnerability(t *testing.T) {
	cd := &common.CoreData{
		Config: &config.RuntimeConfig{
			BaseURL: "http://example.com",
		},
	}
	// Inject a valid session or minimal setup if needed, but SanitizeBackURL
	// primarily relies on config and the input string.
	// However, we need to make sure we don't crash.
	// The function signature is (cd *CoreData) SanitizeBackURL(r *http.Request, raw string) string

	tests := []struct {
		name     string
		raw      string
		expected string // We expect empty string for invalid/unsafe URLs
	}{
		{
			name:     "Protocol-relative URL (Open Redirect)",
			raw:      "//malicious.com",
			expected: "",
		},
		{
			name:     "Absolute URL malicious",
			raw:      "http://malicious.com",
			expected: "",
		},
		{
			name:     "Relative URL valid",
			raw:      "/home",
			expected: "/home",
		},
		{
			name:     "Allowed host",
			raw:      "http://example.com/home",
			expected: "/home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			got := cd.SanitizeBackURL(req, tt.raw)
			if got != tt.expected {
				t.Errorf("SanitizeBackURL(%q) = %q; want %q", tt.raw, got, tt.expected)
			}
		})
	}
}
