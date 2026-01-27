package a4code

import (
	"testing"
)

func TestInvalidTags(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[invalid]", `<span data-start-pos="0" data-end-pos="9">[invalid]</span>`},
		{"[invalid text]", `<span data-start-pos="0" data-end-pos="14">[invalid<span data-start-pos="8" data-end-pos="13"> text</span>]</span>`},
		{"[invalid [b bold]]", `<span data-start-pos="0" data-end-pos="18">[invalid<span data-start-pos="8" data-end-pos="9"> </span><strong data-start-pos="9" data-end-pos="17"><span data-start-pos="11" data-end-pos="16"> bold</span></strong>]</span>`},
		{"[foo=bar]", `<span data-start-pos="0" data-end-pos="9">[foo<span data-start-pos="4" data-end-pos="8">=bar</span>]</span>`},
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
