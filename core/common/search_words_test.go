package common

import (
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestSearchWordsFromRequestCachesAndReturnsCopy(t *testing.T) {
	cd := &CoreData{}
	req := httptest.NewRequest("POST", "/search", strings.NewReader("searchwords=Hello+World"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	words := cd.searchWordsFromRequest(req)
	expected := []string{"Hello", "World"}
	if !reflect.DeepEqual(words, expected) {
		t.Fatalf("searchWordsFromRequest() = %v, want %v", words, expected)
	}

	copied := cd.SearchWords()
	if !reflect.DeepEqual(copied, expected) {
		t.Fatalf("SearchWords() = %v, want %v", copied, expected)
	}

	copied[0] = "changed"
	if cd.cache.searchWords[0] != "Hello" {
		t.Fatalf("SearchWords should return a copy, state was mutated: %v", cd.cache.searchWords)
	}
}
