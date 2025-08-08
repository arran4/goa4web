package common

import "testing"

func TestSectionItemType(t *testing.T) {
	tests := []struct {
		section string
		want    string
	}{
		{"forum", "topic"},
		{"privateforum", "topic"},
	}
	for _, tt := range tests {
		if got := sectionItemType(tt.section); got != tt.want {
			t.Errorf("sectionItemType(%q) = %q; want %q", tt.section, got, tt.want)
		}
	}
}
