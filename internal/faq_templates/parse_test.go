package faq_templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTemplateContent(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedVer     string
		expectedDesc    string
		expectedQ       string
		expectedA       string
		expectError     bool
	}{
		{
			name: "Legacy 2-part template",
			content: `Question
===
Answer`,
			expectedVer:  "1",
			expectedDesc: "",
			expectedQ:    "Question",
			expectedA:    "Answer",
			expectError:  false,
		},
		{
			name: "Legacy 3-part template",
			content: `Description
===
Question
===
Answer`,
			expectedVer:  "1",
			expectedDesc: "Description",
			expectedQ:    "Question",
			expectedA:    "Answer",
			expectError:  false,
		},
		{
			name: "New 4-part template",
			content: `2
===
Description
===
Question
===
Answer`,
			expectedVer:  "2",
			expectedDesc: "Description",
			expectedQ:    "Question",
			expectedA:    "Answer",
			expectError:  false,
		},
		{
			name: "Multiline content",
			content: `2
===
Desc line 1
Desc line 2
===
Question line 1
Question line 2
===
Answer line 1
Answer line 2`,
			expectedVer:  "2",
			expectedDesc: "Desc line 1\nDesc line 2",
			expectedQ:    "Question line 1\nQuestion line 2",
			expectedA:    "Answer line 1\nAnswer line 2",
			expectError:  false,
		},
		{
			name: "Legacy Multiline content",
			content: `Question line 1
Question line 2
===
Answer line 1
Answer line 2`,
			expectedVer:  "1",
			expectedDesc: "",
			expectedQ:    "Question line 1\nQuestion line 2",
			expectedA:    "Answer line 1\nAnswer line 2",
			expectError:  false,
		},
		{
			name:        "Invalid format",
			content:     "Invalid content without separator",
			expectError: true,
		},
		{
			name: "Empty 2-part",
			content: `
===
`,
			expectedVer:  "1",
			expectedDesc: "",
			expectedQ:    "",
			expectedA:    "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, desc, q, a, err := ParseTemplateContent(tt.content)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVer, ver)
				assert.Equal(t, tt.expectedDesc, desc)
				assert.Equal(t, tt.expectedQ, q)
				assert.Equal(t, tt.expectedA, a)
			}
		})
	}
}
