package goa4web

import "testing"

func TestFormatBlobComplex(t *testing.T) {
	input := "\\</pre\\>line1\n#line2\n-line3\n\\*bold\\* text"
	got := formatBlob(input)
	want := "</pre>\nline1<br>\n<ol>\n<li>line2</ol>\n<br>\n<ul>\n<li>line3</ul>\n<br>\n<strong>bold</strong> text"
	if got != want {
		t.Errorf("unexpected output\n got: %q\nwant: %q", got, want)
	}
}

func TestFormatCategoriesSimple(t *testing.T) {
	input := "==Cat==\n#item1\n-item2"
	got := formatCategories(input)
	want := "<ul>\n<a href=\"#Cat\">Cat</a><br><br>\n<br>\n</ul>\n"
	if got != want {
		t.Errorf("unexpected output\n got: %q\nwant: %q", got, want)
	}
}
