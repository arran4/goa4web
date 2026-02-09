package common

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/stretchr/testify/assert"
)

func TestMapLinkURL_AllowedHosts(t *testing.T) {
	tests := []struct {
		name         string
		allowedHosts string
		input        string
		wantSigned   bool
	}{
		{
			name:         "No allowed hosts",
			allowedHosts: "",
			input:        "https://example.com/foo",
			wantSigned:   true,
		},
		{
			name:         "Allowed host match",
			allowedHosts: "example.com",
			input:        "https://example.com/foo",
			wantSigned:   false,
		},
		{
			name:         "Case insensitive allowed host match",
			allowedHosts: "example.com",
			input:        "https://EXAMPLE.COM/foo",
			wantSigned:   false,
		},
		{
			name:         "Allowed host mismatch",
			allowedHosts: "example.org",
			input:        "https://example.com/foo",
			wantSigned:   true,
		},
		{
			name:         "Multiple allowed hosts match first",
			allowedHosts: "example.com,example.org",
			input:        "https://example.com/foo",
			wantSigned:   false,
		},
		{
			name:         "Multiple allowed hosts match second",
			allowedHosts: "example.org,example.com",
			input:        "https://example.com/foo",
			wantSigned:   false,
		},
		{
			name:         "Multiple allowed hosts with spaces",
			allowedHosts: " example.org , example.com ",
			input:        "https://example.com/foo",
			wantSigned:   false,
		},
		{
			name:         "Subdomain mismatch (exact match required)",
			allowedHosts: "example.com",
			input:        "https://sub.example.com/foo",
			wantSigned:   true,
		},
		{
			name:         "Invalid URL",
			allowedHosts: "example.com",
			input:        "://invalid",
			wantSigned:   false, // Not starting with http/https, so returned as is
		},
		{
			name:         "Invalid URL starting with http",
			allowedHosts: "example.com",
			input:        "http://%",
			wantSigned:   true, // url.Parse fails, so proceeds to sign
		},
		{
			name:         "Not http/https",
			allowedHosts: "example.com",
			input:        "ftp://example.com",
			wantSigned:   false, // MapLinkURL returns val directly for non-http/https
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.RuntimeConfig{
				LinkAllowedHosts: tt.allowedHosts,
				LinkSignSecret:   "testsecret",
				BaseURL:          "http://mysite.com",
			}

			cd := &CoreData{
				Config:      cfg,
				LinkSignKey: "testsecret",
			}

			got := cd.MapLinkURL("a", tt.input)

			if tt.wantSigned {
				assert.True(t, strings.Contains(got, "/goto?u="), "Expected signed URL, got: %s", got)
				assert.True(t, strings.Contains(got, "sig="), "Expected signature, got: %s", got)
			} else {
				assert.Equal(t, tt.input, got)
			}
		})
	}
}
