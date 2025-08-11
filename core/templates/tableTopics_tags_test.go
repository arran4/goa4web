package templates

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed site/tableTopics.gohtml
var tableTopicsTemplate string

func TestTableTopicsShowsLabels(t *testing.T) {
	if !strings.Contains(tableTopicsTemplate, "topicLabels") {
		t.Fatalf("tableTopics template missing topicLabels usage")
	}
}
