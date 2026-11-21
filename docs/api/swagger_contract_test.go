package api

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/images"
	"github.com/arran4/goa4web/handlers/privateforum"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"go.yaml.in/yaml/v3"
)

func TestSwaggerContractsMatchPublishedAPIRoutes(t *testing.T) {
	staticSwagger := CanonicalSwaggerYAML()
	validateOpenAPI3Document(t, staticSwagger)

	renderedDocument, err := RenderSwagger("https://example.test/goa", "test")
	if err != nil {
		t.Fatalf("RenderSwagger: %v", err)
	}
	validateOpenAPI3Document(t, renderedDocument)

	static := parseOpenAPI(t, staticSwagger)
	rendered := parseOpenAPI(t, renderedDocument)

	validateOpenAPI(t, static)
	validateOpenAPI(t, rendered)
	assertEquivalentExceptServerAndVersionMetadata(t, static, rendered)

	liveRouter := buildPublishedAPIRouter()
	for _, contract := range []struct {
		name string
		spec map[string]any
	}{
		{name: "static", spec: static},
		{name: "rendered", spec: rendered},
	} {
		t.Run(contract.name, func(t *testing.T) {
			if err := verifyRouterAgainstOpenAPI(liveRouter, contract.spec); err != nil {
				t.Fatalf("router/OpenAPI drift: %v", err)
			}
			assertPublishedPayloadAndScopeContracts(t, contract.spec)
		})
	}
}

func TestRenderSwaggerEscapesUntrustedBaseURL(t *testing.T) {
	tricky := "https://example.test/goa\"\nmalicious: true"
	rendered, err := RenderSwagger(tricky, "1.2.3")
	if err != nil {
		t.Fatalf("RenderSwagger: %v", err)
	}
	validateOpenAPI3Document(t, rendered)
	spec := parseOpenAPI(t, rendered)
	if _, injected := spec["malicious"]; injected {
		t.Fatal("BaseURL injected top-level YAML key")
	}
	servers := spec["servers"].([]any)
	gotURL := stringAt(t, servers[0].(map[string]any), "url")
	if gotURL != tricky {
		t.Fatalf("server URL = %q, want %q", gotURL, tricky)
	}
	if gotVer := stringAt(t, mapAt(t, spec, "info"), "version"); gotVer != "1.2.3" {
		t.Fatalf("info.version = %q, want 1.2.3", gotVer)
	}
}

// TestSwaggerDriftDetectionDetectsLiveRouteMutation verifies that adding or changing
// a live gorilla/mux API route without updating swagger.yaml fails the contract guard.
func TestSwaggerDriftDetectionDetectsLiveRouteMutation(t *testing.T) {
	spec := parseOpenAPI(t, CanonicalSwaggerYAML())

	r := buildPublishedAPIRouter()
	r.HandleFunc("/private/api/undocumented", func(http.ResponseWriter, *http.Request) {}).Methods(http.MethodPost)

	err := verifyRouterAgainstOpenAPI(r, spec)
	if err == nil {
		t.Fatal("expected drift error when undocumented live API route is registered")
	}
	if !strings.Contains(err.Error(), "POST /private/api/undocumented") {
		t.Fatalf("drift error = %v, want mention of POST /private/api/undocumented", err)
	}
}

func buildPublishedAPIRouter() *mux.Router {
	r := mux.NewRouter()
	cfg := &config.RuntimeConfig{}
	privateforum.RegisterRoutes(r, cfg)
	images.RegisterRoutes(r, cfg)
	return r
}

func collectLiveAPIRoutes(r *mux.Router) (map[string]struct{}, error) {
	live := make(map[string]struct{})
	err := r.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil || path == "" {
			return nil
		}
		if !strings.HasPrefix(path, "/private/api") && !strings.HasPrefix(path, "/images/api") {
			return nil
		}
		// /private/api/user-exists is a session-authenticated browser UI helper
		// (RequiresAnAccount), not part of the public bearer-token OpenAPI contract.
		if path == "/private/api/user-exists" {
			return nil
		}
		methods, err := route.GetMethods()
		if err != nil || len(methods) == 0 {
			return nil
		}
		for _, m := range methods {
			live[strings.ToUpper(m)+" "+path] = struct{}{}
		}
		return nil
	})
	return live, err
}

func collectDocumentedAPIRoutes(spec map[string]any) map[string]struct{} {
	documented := make(map[string]struct{})
	paths, _ := spec["paths"].(map[string]any)
	for path, rawOps := range paths {
		ops, ok := rawOps.(map[string]any)
		if !ok {
			continue
		}
		for method := range ops {
			switch strings.ToLower(method) {
			case "get", "post", "put", "patch", "delete", "head", "options":
				documented[strings.ToUpper(method)+" "+path] = struct{}{}
			}
		}
	}
	return documented
}

// verifyRouterAgainstOpenAPI enforces exact (method, path) parity between live
// gorilla/mux routes under /private/api and /images/api and the OpenAPI spec.
// Note: handler-internal x-required-scopes and form/JSON DTO fields cannot be
// extracted from mux.Route alone and are verified separately below.
func verifyRouterAgainstOpenAPI(r *mux.Router, spec map[string]any) error {
	live, err := collectLiveAPIRoutes(r)
	if err != nil {
		return fmt.Errorf("walk live router: %w", err)
	}
	documented := collectDocumentedAPIRoutes(spec)

	var missingInSpec, extraInSpec []string
	for k := range live {
		if _, ok := documented[k]; !ok {
			missingInSpec = append(missingInSpec, k)
		}
	}
	for k := range documented {
		if _, ok := live[k]; !ok {
			extraInSpec = append(extraInSpec, k)
		}
	}
	sort.Strings(missingInSpec)
	sort.Strings(extraInSpec)
	if len(missingInSpec) > 0 || len(extraInSpec) > 0 {
		return fmt.Errorf("missing in OpenAPI: %v; missing in live router: %v", missingInSpec, extraInSpec)
	}
	return nil
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

func assertEquivalentExceptServerAndVersionMetadata(t *testing.T, static, rendered map[string]any) {
	t.Helper()
	staticCopy := cloneSpec(t, static)
	renderedCopy := cloneSpec(t, rendered)
	delete(staticCopy, "servers")
	delete(renderedCopy, "servers")
	delete(mapAt(t, staticCopy, "info"), "version")
	delete(mapAt(t, renderedCopy, "info"), "version")
	if !reflect.DeepEqual(staticCopy, renderedCopy) {
		t.Fatalf("static and downloadable OpenAPI contracts differ outside server/version metadata")
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

func assertPublishedPayloadAndScopeContracts(t *testing.T, spec map[string]any) {
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
		"/private/api/topic/{topic}/thread/{thread}/read":  {method: "post", scope: "private_forum:write"},
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
	server := stringAt(t, spec["servers"].([]any)[0].(map[string]any), "url")
	if server != "https://example.test/goa" && server != "http://localhost:8080" {
		t.Fatalf("server URL = %q, want a site root URL", server)
	}
	assertTopicCreateContract(t, mapAt(t, mapAt(t, paths, "/private/api/topics"), "post"))
	assertReplyContract(t, mapAt(t, mapAt(t, paths, "/private/api/topic/{topic}/thread/{thread}/reply"), "post"))
	assertMarkReadContract(t, mapAt(t, mapAt(t, paths, "/private/api/topic/{topic}/thread/{thread}/read"), "post"))
	comments := mapAt(t, mapAt(t, mapAt(t, mapAt(t, spec, "components"), "schemas"), "Comment"), "properties")
	if got := stringAt(t, mapAt(t, comments, "activity_id"), "$ref"); got != "#/components/schemas/DisplayedActivityID" {
		t.Fatalf("comment activity identifier reference = %q", got)
	}
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
	assertResponseProperties(t, mapAt(t, responses, "200"), "status", "thread_id", "topic_id", "comment_id", "url", "appended", "activity_id", "acknowledgement_failed")
	for _, status := range []string{"400", "401", "403", "404", "500"} {
		if _, ok := responses[status]; !ok {
			t.Fatalf("reply missing %s response", status)
		}
	}
}

func assertMarkReadContract(t *testing.T, operation map[string]any) {
	t.Helper()
	responses := mapAt(t, operation, "responses")
	assertResponseProperties(t, mapAt(t, responses, "200"), "status")
	for _, status := range []string{"400", "401", "403", "404", "500"} {
		if _, ok := responses[status]; !ok {
			t.Fatalf("mark read missing %s response", status)
		}
	}
	form := mapAt(t, mapAt(t, mapAt(t, operation, "requestBody"), "content"), "application/x-www-form-urlencoded")
	schema := mapAt(t, form, "schema")
	if !reflect.DeepEqual(schema["required"], []any{"last_activity"}) {
		t.Fatalf("mark read required fields = %#v", schema["required"])
	}
	properties := mapAt(t, schema, "properties")
	if got := stringAt(t, mapAt(t, properties, "last_activity"), "$ref"); got != "#/components/schemas/DisplayedActivityID" {
		t.Fatalf("mark read activity identifier reference = %q", got)
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
