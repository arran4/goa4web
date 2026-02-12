package a4code

import (
	"testing"
)

func TestInvalidTags(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[invalid]", `<span data-start-pos="0" data-end-pos="0">[invalid]</span>`},
		{"[invalid text]", `<span data-start-pos="0" data-end-pos="4">[invalid<span data-start-pos="0" data-end-pos="4">text</span>]</span>`},
		{"[invalid [b bold]]", `<span data-start-pos="0" data-end-pos="4">[invalid<strong data-start-pos="0" data-end-pos="4"><span data-start-pos="0" data-end-pos="4">bold</span></strong>]</span>`},
		{"[foo=bar]", `<span data-start-pos="0" data-end-pos="3">[foo<span data-start-pos="0" data-end-pos="3">bar</span>]</span>`},
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
