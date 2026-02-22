package email_test

import (
	"net/mail"
	"testing"

	"github.com/arran4/goa4web/internal/email"
)

func TestParseAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected mail.Address
	}{
		{
			input:    "user@example.com",
			expected: mail.Address{Address: "user@example.com"},
		},
		{
			input:    "John Doe <john@example.com>",
			expected: mail.Address{Name: "John Doe", Address: "john@example.com"},
		},
		{
			input:    "invalid-address",
			expected: mail.Address{Address: "invalid-address"},
		},
		{
			input:    "",
			expected: mail.Address{Address: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := email.ParseAddress(tt.input)
			if got.Address != tt.expected.Address || got.Name != tt.expected.Name {
				t.Errorf("ParseAddress(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
