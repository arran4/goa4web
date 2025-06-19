package main

import (
	"net/http/httptest"
	"testing"
)

func TestTemplateFuncsFirstline(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	funcs := NewFuncs(r)
	first := funcs["firstline"].(func(string) string)
	if got := first("a\nb\n"); got != "a" {
		t.Errorf("firstline=%q", got)
	}
}

func TestTemplateFuncsLeft(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	funcs := NewFuncs(r)
	left := funcs["left"].(func(int, string) string)
	if got := left(3, "hello"); got != "hel" {
		t.Errorf("left short=%q", got)
	}
	if got := left(10, "hi"); got != "hi" {
		t.Errorf("left long=%q", got)
	}
}
