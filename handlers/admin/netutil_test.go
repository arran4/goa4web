package admin

import "testing"

func TestNormalizeIPNet(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{" 192.168.0.1 ", "192.168.0.1"},
		{"192.168.0.0/24", "192.168.0.0/24"},
		{"2001:db8::1", "2001:db8::1"},
		{"2001:db8::/32", "2001:db8::/32"},
		{"invalid", "invalid"},
	}
	for _, tt := range tests {
		if got := NormalizeIPNet(tt.in); got != tt.out {
			t.Errorf("NormalizeIPNet(%q)=%q want %q", tt.in, got, tt.out)
		}
	}
}
