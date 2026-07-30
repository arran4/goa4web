package common

import (
	"testing"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		want     string
	}{
		{
			name:     "empty duration",
			duration: "",
			want:     "",
		},
		{
			name:     "raw seconds - minutes",
			duration: "120",
			want:     "2:00",
		},
		{
			name:     "raw seconds - hours",
			duration: "3665",
			want:     "1:01:05",
		},
		{
			name:     "ISO8601 - seconds only",
			duration: "PT45S",
			want:     "0:45",
		},
		{
			name:     "ISO8601 - minutes only",
			duration: "PT1M",
			want:     "1:00",
		},
		{
			name:     "ISO8601 - minutes and seconds",
			duration: "PT2M3S",
			want:     "2:03",
		},
		{
			name:     "ISO8601 - hours, minutes and seconds",
			duration: "PT1H2M3S",
			want:     "1:02:03",
		},
		{
			name:     "ISO8601 - hours only",
			duration: "PT2H",
			want:     "2:00:00",
		},
		{
			name:     "invalid format",
			duration: "invalid",
			want:     "invalid",
		},
		{
			name:     "zero seconds",
			duration: "PT0S",
			want:     "",
		},
		{
			name:     "zero seconds raw",
			duration: "0",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.duration); got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
