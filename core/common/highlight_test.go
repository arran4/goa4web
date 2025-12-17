package common

import (
	"strings"
	"testing"
)

func TestHighlightSearchTermsEscapesAndHighlights(t *testing.T) {
	got := string(HighlightSearchTerms("<b>Cats</b> & Dogs", []string{"cats", "dogs"}))
	want := "&lt;b&gt;<mark>Cats</mark>&lt;/b&gt; &amp; <mark>Dogs</mark>"
	if got != want {
		t.Fatalf("unexpected highlight output:\n got: %s\nwant: %s", got, want)
	}
}

func TestHighlightSearchTermsRespectsWordBoundaries(t *testing.T) {
	text := "concatenate co-operate's cat"
	got := string(HighlightSearchTerms(text, []string{"cat", "co-operate's"}))
	if strings.Contains(got, "<mark>concatenate</mark>") {
		t.Fatalf("unexpected highlight in unrelated word: %s", got)
	}
	if !strings.Contains(got, "<mark>co-operate&#39;s</mark>") {
		t.Fatalf("expected highlight for punctuation word: %s", got)
	}
	if !strings.Contains(got, "<mark>cat</mark>") {
		t.Fatalf("expected highlight for standalone word: %s", got)
	}
}

func TestHighlightSearchTermsWithoutWordsEscapesHTML(t *testing.T) {
	got := string(HighlightSearchTerms("<b>cat</b>", nil))
	if got != "&lt;b&gt;cat&lt;/b&gt;" {
		t.Fatalf("expected escaped HTML when no search words: %s", got)
	}
}
