package goa4web

import "testing"

func TestBreakupTextToWordsEdge(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"..hi--there--foo", []string{"hi--there--foo"}},
		{"end.", []string{"end"}},
		{"it's foo", []string{"it's", "foo"}},
		{"---abc", []string{"---abc"}},
	}
	for _, c := range cases {
		got := breakupTextToWords(c.in)
		if len(got) != len(c.want) {
			t.Errorf("%q len=%d want %d", c.in, len(got), len(c.want))
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q index %d got %q want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestIsAlphanumericOrPunctuationExtra(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'ñ', true},
		{'?', false},
		{'-', true},
		{'é', true},
	}
	for _, tt := range tests {
		if got := isAlphanumericOrPunctuation(tt.r); got != tt.want {
			t.Errorf("%q got %v want %v", string(tt.r), got, tt.want)
		}
	}
}
