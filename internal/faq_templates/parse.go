package faq_templates

import (
	"fmt"
	"strings"
)

// ParseTemplateContent splits a template into question and answer parts, and optionally a description.
func ParseTemplateContent(content string) (string, string, string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	parts := strings.SplitN(content, "\n===\n", 3)
	if len(parts) == 3 {
		description := strings.TrimSpace(parts[0])
		question := strings.TrimSpace(parts[1])
		answer := strings.TrimSpace(parts[2])
		return description, question, answer, nil
	} else if len(parts) == 2 {
		question := strings.TrimSpace(parts[0])
		answer := strings.TrimSpace(parts[1])
		return "", question, answer, nil
	}
	return "", "", "", fmt.Errorf("invalid template format: missing '===' separator")
}
