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
			name:          "defaults",
			args:          []string{},
			expected:      AsOptions{Extended: false},
			expectedError: false,
		},
		{
			name:          "extended flag",
			args:          []string{"-extended"},
			expected:      AsOptions{Extended: true},
			expectedError: false,
		},
		{
			name:          "extended flag explicit true",
			args:          []string{"-extended=true"},
			expected:      AsOptions{Extended: true},
			expectedError: false,
		},
		{
			name:          "extended flag explicit false",
			args:          []string{"-extended=false"},
			expected:      AsOptions{Extended: false},
			expectedError: false,
		},
		{
			name:          "unknown flag",
			args:          []string{"-unknown"},
			expected:      AsOptions{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
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
