package faq_templates

import (
	"fmt"
	"strings"
)

// ParseTemplateContent splits a template into version, description, question and answer parts.
func ParseTemplateContent(content string) (string, string, string, string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	parts := strings.SplitN(content, "\n===\n", 4)
	if len(parts) == 4 {
		version := strings.TrimSpace(parts[0])
		description := strings.TrimSpace(parts[1])
		question := strings.TrimSpace(parts[2])
		answer := strings.TrimSpace(parts[3])
		return version, description, question, answer, nil
	} else if len(parts) == 3 {
		description := strings.TrimSpace(parts[0])
		question := strings.TrimSpace(parts[1])
		answer := strings.TrimSpace(parts[2])
		return "1", description, question, answer, nil
	} else if len(parts) == 2 {
		question := strings.TrimSpace(parts[0])
		answer := strings.TrimSpace(parts[1])
		return "1", "", question, answer, nil
	}
	return "", "", "", "", fmt.Errorf("invalid template format: missing '===' separator")
}
