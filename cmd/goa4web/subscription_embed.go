package main

import (
	"embed"
	"fmt"
)

//go:embed subscription_templates/*.txt
var embeddedSubscriptionTemplates embed.FS

func getEmbeddedTemplate(name string) ([]byte, error) {
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
