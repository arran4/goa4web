package a4code_test

import (
	"testing"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/a4code/ast"
	"github.com/stretchr/testify/assert"
)

func TestCodeBlockEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Nested escaped bracket",
			input:    `[code [quote test\]]`,
			expected: `[quote test]`,
		},
		{
			name:     "Escaped bracket prevents termination (with space)",
			input:    `[code C:\]path ]`,
			expected: `C:]path `,
		},
        {
            name:     "Escaped bracket prevents termination (EOF case)",
            input:    `[code C:\]path]`,
            expected: `C:]path`, // Captures C:]path, last bracket terminates block
        },
		{
			name:     "Standard block content",
			input:    `[code [b]bold[/b]]`,
			expected: `[b]bold[/b]`,
		},
		{
			name:     "Literal bracket at end",
			input:    `[code smile :-\] ]`,
			expected: `smile :-] `,
		},
		{
			name:     "Multiple nested escaped brackets",
			input:    `[code [ [ ] \] ]`,
			expected: `[ [ ] ] `,
		},
		{
			name:     "Escaped open bracket literal",
			input:    `[code \[literal]`,
			expected: `[literal`,
		},
        {
            name:     "Escaped open bracket literal closed",
            input:    `[code \[literal\]]`,
            expected: `[literal]`,
        },
        {
            name: "New line handling",
            input: "[code \nline1\nline2\n]",
            expected: "line1\nline2\n", // Leading newline consumed by parser
        },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := a4code.ParseString(tc.input)
			assert.NoError(t, err)
			assert.NotEmpty(t, root.Children)
			codeNode, ok := root.Children[0].(*ast.Code)
			assert.True(t, ok, "Expected ast.Code node")
			assert.Equal(t, tc.expected, codeNode.Value)
		})
	}
}
