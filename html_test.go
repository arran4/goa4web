package main

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
