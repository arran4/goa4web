package main

import "testing"

func TestIsAlphanumericOrPunctuation(t *testing.T) {
	cases := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'-', true},
		{'\'', true},
		{'!', false},
	}
	for _, c := range cases {
		if got := isAlphanumericOrPunctuation(c.r); got != c.want {
			t.Errorf("%q got %v want %v", string(c.r), got, c.want)
		}
	}
}

func TestBreakupTextToWords(t *testing.T) {
	in := "Hello, world! It's-me"
	words := breakupTextToWords(in)
	want := []string{"Hello", "world", "It's-me"}
	if len(words) != len(want) {
		t.Fatalf("len=%d want %d", len(words), len(want))
	}
	for i := range words {
		if words[i] != want[i] {
			t.Errorf("index %d got %q want %q", i, words[i], want[i])
		}
	}
}
