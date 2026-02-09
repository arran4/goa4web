package configformat

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAsFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expected      AsOptions
		expectedError bool
	}{
		{
			name:     "default values",
			args:     []string{},
			expected: AsOptions{Extended: false},
		},
		{
			name:     "extended flag",
			args:     []string{"-extended"},
			expected: AsOptions{Extended: true},
		},
		{
			name:     "extended flag long",
			args:     []string{"--extended"},
			expected: AsOptions{Extended: true},
		},
		{
			name:          "unknown flag",
			args:          []string{"-unknown"},
			expectedError: true,
		},
		{
			name:     "explicit false",
			args:     []string{"-extended=false"},
			expected: AsOptions{Extended: false},
		},
		{
			name:     "explicit true",
			args:     []string{"-extended=true"},
			expected: AsOptions{Extended: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			// Suppress output for expected errors
			fs.SetOutput(io.Discard)
			opts, err := ParseAsFlags(fs, tt.args)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, opts)
			}
		})
	}
}
