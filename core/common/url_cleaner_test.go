package common

import (
	"testing"
)

func TestCanonicalizeExternalURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
	}{
		{
			name: "Removes utm tags",
			raw:  "https://example.com/?utm_source=google&utm_medium=cpc&id=123",
			want: "https://example.com/?id=123",
		},
		{
			name: "Retains functional tags",
			raw:  "https://example.com/?accessToken=abc&resource=xyz&X-Amz-Security-Token=foo",
			want: "https://example.com/?accessToken=abc&resource=xyz&X-Amz-Security-Token=foo",
		},
		{
			name: "Mixed tags",
			raw:  "https://example.com/?accessToken=abc&utm_campaign=sale",
			want: "https://example.com/?accessToken=abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanonicalizeExternalURL(tt.raw); got != tt.want {
				t.Errorf("CanonicalizeExternalURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateURLHash(t *testing.T) {
	hash1 := GenerateURLHash("http://example.com")
	hash2 := GenerateURLHash("http://example.com")
	if hash1 != hash2 {
		t.Errorf("Hashes must be deterministic")
	}
}
