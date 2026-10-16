package privateforum

import (
	"bytes"
	"os"

	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// SwaggerSpec is a simplified structure to parse the OpenAPI yaml document
type SwaggerSpec struct {
	Paths map[string]map[string]struct {
		Summary     string `yaml:"summary"`
		Description string `yaml:"description"`
		Parameters  []struct {
			Name   string `yaml:"name"`
			In     string `yaml:"in"`
			Schema struct {
				Type    string `yaml:"type"`
				Minimum int    `yaml:"minimum"`
				Maximum int    `yaml:"maximum"`
			} `yaml:"schema"`
		} `yaml:"parameters"`
		RequestBody struct {
			Content map[string]struct {
				Schema struct {
					Properties map[string]interface{} `yaml:"properties"`
					Required   []string               `yaml:"required"`
				} `yaml:"schema"`
			} `yaml:"content"`
		} `yaml:"requestBody"`
		Responses map[string]struct {
			Description string `yaml:"description"`
			Content     map[string]struct {
				Schema struct {
					Properties map[string]interface{} `yaml:"properties"`
					Ref        string                 `yaml:"$ref"`
				} `yaml:"schema"`
			} `yaml:"content"`
		} `yaml:"responses"`
	} `yaml:"paths"`
	Components struct {
		Schemas map[string]interface{} `yaml:"schemas"`
	} `yaml:"components"`
}

func TestSwaggerDefinitions(t *testing.T) {
	// Paths to verify
	staticPath := "../../docs/api/swagger.yaml"
	templatePath := "../../core/templates/site/domains/user/swagger.gohtml"

	// 1. Read static yaml
	staticContent, err := os.ReadFile(staticPath)
	require.NoError(t, err, "failed to read static swagger")

	// 2. Read and execute template yaml
	tmpl, err := template.ParseFiles(templatePath)
	require.NoError(t, err, "failed to parse template swagger")
	var tmplBuf bytes.Buffer
	err = tmpl.Execute(&tmplBuf, map[string]interface{}{
		"BaseURL": "http://localhost:8080",
		"Version": "test",
	})
	require.NoError(t, err, "failed to execute template swagger")
	tmplContent := tmplBuf.Bytes()

	// Verify both specs
	verifySwaggerSpec(t, staticContent, "static")
	verifySwaggerSpec(t, tmplContent, "template")
}

func verifySwaggerSpec(t *testing.T, content []byte, source string) {
	var spec SwaggerSpec
	err := yaml.Unmarshal(content, &spec)
	require.NoError(t, err, "failed to unmarshal yaml for %s", source)

	// Ensure error schema exists
	require.Contains(t, spec.Components.Schemas, "ErrorResponse", "ErrorResponse missing in %s", source)

	// Verify Routes
	t.Run("topics-get-"+source, func(t *testing.T) {
		op, ok := spec.Paths["/private/api/topics"]["get"]
		require.True(t, ok, "GET /private/api/topics missing")

		// Check description mentions scopes
		require.Contains(t, op.Description, "private_forum:read")

		// Check responses
		require.Contains(t, op.Responses, "200")
		require.Contains(t, op.Responses, "400")
		require.Contains(t, op.Responses, "401")

		// Check DTO properties
		props := op.Responses["200"].Content["application/json"].Schema.Properties
		require.Contains(t, props, "topics")

		// Quick check on error ref
		require.Equal(t, "#/components/schemas/ErrorResponse", op.Responses["400"].Content["application/json"].Schema.Ref)
	})

	t.Run("topics-post-"+source, func(t *testing.T) {
		op, ok := spec.Paths["/private/api/topics"]["post"]
		require.True(t, ok, "POST /private/api/topics missing")

		require.Contains(t, op.Description, "private_forum:write")
		require.Contains(t, op.Description, "privateforum:topic:create")

		// Check body schema required fields
		schema := op.RequestBody.Content["application/x-www-form-urlencoded"].Schema
		require.Contains(t, schema.Required, "title")
		require.Contains(t, schema.Required, "participants")
		require.NotContains(t, schema.Required, "subject")
		require.Contains(t, schema.Properties, "description")

		// Check responses
		require.Contains(t, op.Responses, "200")
		require.Contains(t, op.Responses["200"].Content["application/json"].Schema.Properties, "topicID")
		require.Contains(t, op.Responses, "400")
		require.Contains(t, op.Responses, "403")
	})

	t.Run("threads-get-"+source, func(t *testing.T) {
		op, ok := spec.Paths["/private/api/topic/{topic}/threads"]["get"]
		require.True(t, ok, "GET threads missing")

		require.Contains(t, op.Description, "private_forum:read")

		// Check bounds on ID
		for _, param := range op.Parameters {
			if param.Name == "topic" && param.In == "path" {
				require.Equal(t, "integer", param.Schema.Type)
				require.Equal(t, 1, param.Schema.Minimum)
				require.Equal(t, 2147483647, param.Schema.Maximum)
			}
		}

		require.Contains(t, op.Responses, "200")
		require.Contains(t, op.Responses, "400")
		require.Contains(t, op.Responses, "404")
	})

	t.Run("thread-reply-post-"+source, func(t *testing.T) {
		op, ok := spec.Paths["/private/api/topic/{topic}/thread/{thread}/reply"]["post"]
		require.True(t, ok, "POST reply missing")

		require.Contains(t, op.Description, "private_forum:write")

		// Check correct reply structure
		props := op.Responses["200"].Content["application/json"].Schema.Properties
		require.Contains(t, props, "status")
		require.Contains(t, props, "thread_id")
		require.Contains(t, props, "topic_id")
		require.Contains(t, props, "comment_id")
		require.Contains(t, props, "url")
		require.Contains(t, props, "appended")

		require.Contains(t, op.Responses, "400")
		require.Contains(t, op.Responses, "404")
	})
}
