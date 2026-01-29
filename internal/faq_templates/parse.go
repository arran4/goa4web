package faq_templates

import (
	"fmt"
	"strings"
)

// ParseTemplateContent splits a template into question and answer parts.
func ParseTemplateContent(content string) (string, string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	parts := strings.SplitN(content, "\n===\n", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid template format: missing '===' separator")
	}
	question := strings.TrimSpace(parts[0])
	answer := strings.TrimSpace(parts[1])
	return question, answer, nil
}
