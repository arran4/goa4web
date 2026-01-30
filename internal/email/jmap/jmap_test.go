package jmap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestGetJMAPDiscoveryEndpoint(t *testing.T) {
	tests := []struct {
		name          string
		override      string
		endpoint      string
		expected      string
		expectError   bool
		expectedError string
	}{
		{
			name:     "Discovery Precedence",
			override: "https://override.com/jmap",
			endpoint: "https://default.com/jmap",
			expected: "https://default.com/jmap",
		},
		{
			name:     "Override Fallback",
			override: "https://override.com/jmap",
			endpoint: "",
			expected: "https://override.com/jmap",
		},
		{
			name:     "Default Endpoint",
			endpoint: "https://default.com/jmap",
			expected: "https://default.com/jmap",
		},
		{
			name:          "No Endpoint Configured",
			expectError:   true,
			expectedError: "email disabled: JMAP_ENDPOINT or JMAP_ENDPOINT_OVERRIDE not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.RuntimeConfig{
				EmailJMAPEndpoint:         tt.endpoint,
				EmailJMAPEndpointOverride: tt.override,
			}
			ep, err := getJMAPDiscoveryEndpoint(cfg)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', but got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				if ep != tt.expected {
					t.Errorf("Expected endpoint '%s', but got '%s'", tt.expected, ep)
				}
			}
		})
	}
}

func TestResolveJMAPSettings(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jmap" {
			session := &SessionResponse{
				APIURL: fmt.Sprintf("%s/jmap", server.URL),
				PrimaryAccounts: map[string]string{
					mailCapabilityURN: "acc1",
				},
				DefaultIdentity: map[string]string{
					mailCapabilityURN: "id1",
				},
			}
			json.NewEncoder(w).Encode(session)
		} else if r.URL.Path == "/jmap" {
			// Mock Identity/get if needed, but session provides it here so it shouldn't be called for identity
			http.Error(w, "Not Expecting API Call", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	cfg := &config.RuntimeConfig{
		EmailJMAPUser:             "user",
		EmailJMAPPass:             "pass",
		EmailJMAPDiscoveryRetries: 1,
	}

	httpClient := server.Client()

	acc, id, ep, err := resolveJMAPSettings(context.Background(), httpClient, cfg, server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if acc != "acc1" {
		t.Errorf("Expected account ID 'acc1', got '%s'", acc)
	}
	if id != "id1" {
		t.Errorf("Expected identity ID 'id1', got '%s'", id)
	}
	if ep != fmt.Sprintf("%s/jmap", server.URL) {
		t.Errorf("Expected endpoint '%s/jmap', got '%s'", server.URL, ep)
	}

	// Test override
	overrideCfg := &config.RuntimeConfig{
		EmailJMAPUser:             "user",
		EmailJMAPPass:             "pass",
		EmailJMAPDiscoveryRetries: 1,
		EmailJMAPEndpointOverride: "https://override.com/api",
	}
	acc, id, ep, err = resolveJMAPSettings(context.Background(), httpClient, overrideCfg, server.URL)
	if err != nil {
		t.Fatalf("Expected no error with override, got %v", err)
	}
	if ep != "https://override.com/api" {
		t.Errorf("Expected override endpoint 'https://override.com/api', got '%s'", ep)
	}
}

func TestProviderFromConfig(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jmap" {
			session := &SessionResponse{
				APIURL: fmt.Sprintf("%s/jmap", server.URL),
				PrimaryAccounts: map[string]string{
					mailCapabilityURN: "acc1",
				},
			}
			json.NewEncoder(w).Encode(session)
		} else if r.URL.Path == "/jmap" {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			methodCalls, ok := req["methodCalls"].([]interface{})
			if !ok || len(methodCalls) == 0 {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			methodCall, ok := methodCalls[0].([]interface{})
			if !ok || len(methodCall) < 2 {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			methodName, ok := methodCall[0].(string)
			if !ok {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if methodName == "Identity/get" {
				resp := map[string]interface{}{
					"methodResponses": [][]interface{}{
						{
							"Identity/get",
							map[string]interface{}{
								"list": []interface{}{
									map[string]interface{}{
										"id": "id1",
									},
								},
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			}
		}
	}))
	defer server.Close()

	// Test case 1: Successful provider creation with discovery
	cfg := &config.RuntimeConfig{
		EmailJMAPEndpoint:         server.URL,
		EmailJMAPDiscoveryRetries: 1,
	}
	provider, err := providerFromConfig(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	p := provider.(*Provider)
	if p.GetAccountID() != "acc1" {
		t.Errorf("Expected account ID 'acc1', got '%s'", p.GetAccountID())
	}
	if p.GetIdentity() != "id1" {
		t.Errorf("Expected identity ID 'id1', got '%s'", p.GetIdentity())
	}
	if p.GetEndpoint() != fmt.Sprintf("%s/jmap", server.URL) {
		t.Errorf("Expected endpoint '%s/jmap', got '%s'", server.URL, p.GetEndpoint())
	}

	// Test case 2: Successful provider creation with manual config
	cfg = &config.RuntimeConfig{
		EmailJMAPEndpoint: server.URL,
		EmailJMAPAccount:  "manual_acc",
		EmailJMAPIdentity: "manual_id",
	}
	provider, err = providerFromConfig(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	p = provider.(*Provider)
	if p.GetAccountID() != "manual_acc" {
		t.Errorf("Expected account ID 'manual_acc', got '%s'", p.GetAccountID())
	}
	if p.GetIdentity() != "manual_id" {
		t.Errorf("Expected identity ID 'manual_id', got '%s'", p.GetIdentity())
	}

	// Test case 3: Successful provider creation with manual config AND override
	cfg = &config.RuntimeConfig{
		EmailJMAPEndpoint:         server.URL,
		EmailJMAPAccount:          "manual_acc",
		EmailJMAPIdentity:         "manual_id",
		EmailJMAPEndpointOverride: "https://override.com/api",
	}
	provider, err = providerFromConfig(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	p = provider.(*Provider)
	if p.GetEndpoint() != "https://override.com/api" {
		t.Errorf("Expected override endpoint 'https://override.com/api', got '%s'", p.GetEndpoint())
	}
}
