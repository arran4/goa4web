package faq_templates

import (
	"testing"
)

func TestLabelsTemplateExists(t *testing.T) {
	names, err := List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	found := false
	for _, name := range names {
		if name == "labels" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'labels' in template list, got %v", names)
	}

	content, err := Get("labels")
	if err != nil {
		t.Fatalf("Get('labels') failed: %v", err)
	}

	if content == "" {
		t.Error("Expected content for 'labels' template, got empty string")
	}
}
