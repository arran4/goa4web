package goa4web

import "testing"

func TestAnchorLink(t *testing.T) {
	got := anchorLink("section", "name")
	want := "<a href=\"#section\">name</a><br>"
	if got != want {
		t.Errorf("anchorLink=%q", got)
	}
}

func TestPageLink(t *testing.T) {
	got := pageLink("home", "Home")
	want := "<a href=\"?page=home\">Home</a><br>"
	if got != want {
		t.Errorf("pageLink=%q", got)
	}
}

func TestCategoryLevel(t *testing.T) {
	if got := categoryLevel("cat", 1); got != "<p><a name=\"cat\"><span style=\"font-size: 16;\">cat</span></a><br>\n" {
		t.Errorf("categoryLevel=%q", got)
	}
	if got := categoryLevel("cat", 3); got != "<p><a name=\"cat\"><span style=\"\">cat</span></a><br>\n" {
		t.Errorf("categoryLevel default=%q", got)
	}
}

func TestExternalLink(t *testing.T) {
	got := externalLink("http://x", "X")
	want := "<a href=\"http://x\" target=\"_blank\">X</a>"
	if got != want {
		t.Errorf("externalLink=%q", got)
	}
}

func TestExternalLinkBadURL(t *testing.T) {
	got := externalLink("javascript:alert(1)", "X")
	want := "javascript:alert(1)"
	if got != want {
		t.Errorf("externalLink bad=%q", got)
	}
}

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
