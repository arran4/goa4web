// Package api embeds the canonical OpenAPI 3 specification and renders deployment-specific copies.
package api

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
)

// canonicalSwaggerYAML is the single source-of-truth OpenAPI 3 specification.
//
//go:embed swagger.yaml
var canonicalSwaggerYAML []byte

// CanonicalSwaggerYAML returns a copy of the checked-in canonical OpenAPI 3 YAML.
func CanonicalSwaggerYAML() []byte {
	return append([]byte(nil), canonicalSwaggerYAML...)
}

// RenderSwagger returns a valid OpenAPI 3 YAML copy derived from the canonical
// specification with only runtime server metadata and version updated.
func RenderSwagger(baseURL, version string) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(canonicalSwaggerYAML, &doc); err != nil {
		return nil, fmt.Errorf("parse canonical swagger: %w", err)
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("canonical swagger root is not a YAML mapping")
	}
	root := doc.Content[0]

	trimmedBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedBase == "" {
		trimmedBase = "/"
	}

	if v := strings.TrimSpace(version); v != "" {
		if infoNode := findMapValue(root, "info"); infoNode != nil && infoNode.Kind == yaml.MappingNode {
			setMapScalar(infoNode, "version", v)
		}
	}

	serverNode := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "url"},
			{Kind: yaml.ScalarNode, Value: trimmedBase},
			{Kind: yaml.ScalarNode, Value: "description"},
			{Kind: yaml.ScalarNode, Value: "Current Server"},
		},
	}
	setMapNode(root, "servers", &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: []*yaml.Node{serverNode},
	})

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return nil, fmt.Errorf("encode rendered swagger: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("close swagger yaml encoder: %w", err)
	}
	return buf.Bytes(), nil
}

func findMapValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func setMapScalar(m *yaml.Node, key, value string) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = &yaml.Node{Kind: yaml.ScalarNode, Value: value}
			return
		}
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value},
	)
}

func setMapNode(m *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = value
			return
		}
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		value,
	)
}
