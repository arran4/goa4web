package api

import (
	"bytes"
	_ "embed"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/arran4/goa4web/core/templates"
	"github.com/getkin/kin-openapi/openapi3"
	"go.yaml.in/yaml/v3"
)

// staticSwagger is the checked-in public API contract.
//
//go:embed swagger.yaml
var staticSwagger []byte

func TestSwaggerContractsMatchPublishedAPIRoutes(t *testing.T) {
	validateOpenAPI3Document(t, staticSwagger)
	renderedDocument := renderSwagger(t, "https://example.test/goa")
	validateOpenAPI3Document(t, renderedDocument)
	static := parseOpenAPI(t, staticSwagger)
	rendered := parseOpenAPI(t, renderedDocument)

	validateOpenAPI(t, static)
	validateOpenAPI(t, rendered)
	assertEquivalentExceptServerMetadata(t, static, rendered)

	for _, contract := range []struct {
		name string
		spec map[string]any
	}{
		{name: "static", spec: static},
		{name: "rendered", spec: rendered},
	} {
		t.Run(contract.name, func(t *testing.T) {
			assertPublishedRouteContract(t, contract.spec)
		})
	}
}

func renderSwagger(t *testing.T, baseURL string) []byte {
	t.Helper()
	var rendered bytes.Buffer
	tmpl, err := template.New("swagger").Parse(string(templates.GetUserSwaggerTemplateData()))
	if err != nil {
		t.Fatalf("parse downloadable Swagger template: %v", err)
	}
	err = tmpl.ExecuteTemplate(&rendered, "user/swagger.gohtml", struct {
		BaseURL string
		Version string
	}{
		BaseURL: baseURL,
		Version: "test",
	})
	if err != nil {
		t.Fatalf("render downloadable Swagger: %v", err)
	}
	return rendered.Bytes()
}

func parseOpenAPI(t *testing.T, document []byte) map[string]any {
	t.Helper()
	var spec map[string]any
	if err := yaml.Unmarshal(document, &spec); err != nil {
		t.Fatalf("parse OpenAPI YAML: %v", err)
	}
	return spec
}

func validateOpenAPI(t *testing.T, spec map[string]any) {
	t.Helper()
	if version := stringAt(t, spec, "openapi"); !strings.HasPrefix(version, "3.") {
		t.Fatalf("openapi version = %q, want 3.x", version)
	}
	if stringAt(t, mapAt(t, spec, "info"), "title") == "" || stringAt(t, mapAt(t, spec, "info"), "version") == "" {
		t.Fatal("OpenAPI info must include title and version")
	}
	servers, ok := spec["servers"].([]any)
	if !ok || len(servers) != 1 || stringAt(t, servers[0].(map[string]any), "url") == "" {
		t.Fatal("OpenAPI must define exactly one non-empty server URL")
	}
	if _, ok := spec["paths"].(map[string]any); !ok {
		t.Fatal("OpenAPI must define paths")
	}
}

func validateOpenAPI3Document(t *testing.T, document []byte) {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(document)
	if err != nil {
		t.Fatalf("load OpenAPI 3 document: %v", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		t.Fatalf("validate OpenAPI 3 document: %v", err)
	}
}

func assertEquivalentExceptServerMetadata(t *testing.T, static, rendered map[string]any) {
	t.Helper()
	staticCopy := cloneSpec(t, static)
	renderedCopy := cloneSpec(t, rendered)
	delete(staticCopy, "servers")
	delete(renderedCopy, "servers")
	if !reflect.DeepEqual(staticCopy, renderedCopy) {
		t.Fatalf("static and downloadable OpenAPI contracts differ outside server metadata")
	}
}

func cloneSpec(t *testing.T, spec map[string]any) map[string]any {
	t.Helper()
	data, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal OpenAPI clone: %v", err)
	}
	return parseOpenAPI(t, data)
}

func assertPublishedRouteContract(t *testing.T, spec map[string]any) {
	t.Helper()
	paths := mapAt(t, spec, "paths")
	for path, requirement := range map[string]struct {
		method string
		scope  string
	}{
		"/private/api/topics":                              {method: "get", scope: "private_forum:read"},
		"/private/api/topics#post":                         {method: "post", scope: "private_forum:write"},
		"/private/api/topic/{topic}/threads":               {method: "get", scope: "private_forum:read"},
		"/private/api/topic/{topic}/thread/{thread}":       {method: "get", scope: "private_forum:read"},
		"/private/api/topic/{topic}/thread/{thread}/reply": {method: "post", scope: "private_forum:write"},
		"/images/api":        {method: "get", scope: "images:read"},
		"/images/api/upload": {method: "post", scope: "images:write"},
	} {
		pathName := strings.TrimSuffix(path, "#post")
		operations := mapAt(t, paths, pathName)
		operation := mapAt(t, operations, requirement.method)
		if scopes, ok := operation["x-required-scopes"].([]any); !ok || len(scopes) != 1 || scopes[0] != requirement.scope {
			t.Fatalf("%s %s required scopes = %#v, want %q", strings.ToUpper(requirement.method), pathName, operation["x-required-scopes"], requirement.scope)
		}
	}
	if len(paths) != 6 {
		t.Fatalf("paths = %d, want exactly the six published API paths", len(paths))
	}
	server := stringAt(t, spec["servers"].([]any)[0].(map[string]any), "url")
	if server != "https://example.test/goa" && server != "http://localhost:8080" {
		t.Fatalf("server URL = %q, want a site root URL", server)
	}
	assertTopicCreateContract(t, mapAt(t, mapAt(t, paths, "/private/api/topics"), "post"))
	assertReplyContract(t, mapAt(t, mapAt(t, paths, "/private/api/topic/{topic}/thread/{thread}/reply"), "post"))
}

func assertTopicCreateContract(t *testing.T, operation map[string]any) {
	t.Helper()
	schema := mapAt(t, mapAt(t, mapAt(t, operation, "requestBody"), "content"), "application/x-www-form-urlencoded")
	properties := mapAt(t, mapAt(t, schema, "schema"), "properties")
	if _, ok := properties["title"]; !ok {
		t.Fatal("topic creation must require title")
	}
	participants := mapAt(t, properties, "participants")
	if stringAt(t, participants, "type") != "array" {
		t.Fatal("topic creation participants must be an array")
	}
	if _, ok := properties["subject"]; ok {
		t.Fatal("topic creation must not document subject")
	}
	responses := mapAt(t, operation, "responses")
	assertResponseProperties(t, mapAt(t, responses, "200"), "status", "message", "topicID")
	for _, status := range []string{"400", "401", "403", "500"} {
		if _, ok := responses[status]; !ok {
			t.Fatalf("topic creation missing %s response", status)
		}
	}
}

func assertReplyContract(t *testing.T, operation map[string]any) {
	t.Helper()
	responses := mapAt(t, operation, "responses")
	assertResponseProperties(t, mapAt(t, responses, "200"), "status", "thread_id", "topic_id", "comment_id", "url", "appended")
	for _, status := range []string{"400", "401", "403", "404", "500"} {
		if _, ok := responses[status]; !ok {
			t.Fatalf("reply missing %s response", status)
		}
	}
}

func assertResponseProperties(t *testing.T, response map[string]any, names ...string) {
	t.Helper()
	content := mapAt(t, response, "content")
	properties := mapAt(t, mapAt(t, mapAt(t, content, "application/json"), "schema"), "properties")
	for _, name := range names {
		if _, ok := properties[name]; !ok {
			t.Fatalf("response missing %q", name)
		}
	}
}

func mapAt(t *testing.T, value map[string]any, key string) map[string]any {
	t.Helper()
	got, ok := value[key].(map[string]any)
	if !ok {
		t.Fatalf("%q is not an object", key)
	}
	return got
}

func stringAt(t *testing.T, value map[string]any, key string) string {
	t.Helper()
	got, ok := value[key].(string)
	if !ok {
		t.Fatalf("%q is not a string", key)
	}
	return got
}
