package subscriptiontemplates

import (
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed subscription_templates/*.txt
var embeddedSubscriptionTemplates embed.FS

// Pattern represents a parsed subscription template pattern.
type Pattern struct {
	Method  string
	Pattern string
}

// GetEmbeddedTemplate returns the contents of an embedded subscription template by name.
func GetEmbeddedTemplate(name string) ([]byte, error) {
	// Try direct match
	data, err := embeddedSubscriptionTemplates.ReadFile(name)
	if err == nil {
		return data, nil
	}
	// Try with directory prefix
	data, err = embeddedSubscriptionTemplates.ReadFile("subscription_templates/" + name)
	if err == nil {
		return data, nil
	}
	// Try with .txt extension
	data, err = embeddedSubscriptionTemplates.ReadFile("subscription_templates/" + name + ".txt")
	if err == nil {
		return data, nil
	}
	return nil, fmt.Errorf("template %s not found in embedded FS", name)
}

// ListEmbeddedTemplates returns the available embedded template names.
func ListEmbeddedTemplates() ([]string, error) {
	entries, err := embeddedSubscriptionTemplates.ReadDir("subscription_templates")
	if err != nil {
		return nil, fmt.Errorf("read embedded templates: %w", err)
	}
	items := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".txt")
		if name == "" {
			continue
		}
		items = append(items, name)
	}
	sort.Strings(items)
	return items, nil
}

// ParseTemplatePatterns parses subscription template content into patterns.
func ParseTemplatePatterns(content string) []Pattern {
	lines := splitLines(content)
	patterns := make([]Pattern, 0, len(lines))
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}
		method := "internal"
		pattern := line
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			m := strings.ToLower(parts[0])
			if m == "email" || m == "internal" {
				method = m
				pattern = parts[1]
			}
		}
		patterns = append(patterns, Pattern{Method: method, Pattern: pattern})
	}
	return patterns
}

func splitLines(s string) []string {
	var lines []string
	var line string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, line)
			line = ""
		} else {
			line += string(r)
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}
