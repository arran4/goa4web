package a4code

import (
	"testing"
)

func TestInvalidTags(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[invalid]", "[invalid]"},
		{"[invalid text]", "[invalid text]"},
		{"[invalid [b bold]]", "[invalid <strong> bold</strong>]"},
		{"[foo=bar]", "[foo=bar]"},
	}

	for _, tc := range tests {
		node, err := ParseString(tc.input)
		if err != nil {
			t.Fatalf("ParseString(%q) error: %v", tc.input, err)
		}
		output := ToHTML(node)
		if output != tc.expected {
			t.Errorf("Input: %q\nExpected: %q\nGot:      %q", tc.input, tc.expected, output)
		}
	}
}
