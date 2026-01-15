package jmap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestProviderFromConfigDiscoversSession(t *testing.T) {
	t.Parallel()

	var sawWellKnown bool
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// New logic respects custom paths, so we expect /jmap if that's what we configured
		if r.URL.Path != "/.well-known/jmap" && r.URL.Path != "/jmap" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "user" || pass != "pass" {
			t.Fatalf("unexpected auth: %s %s %v", user, pass, ok)
		}
		sawWellKnown = true
		resp := SessionResponse{
			APIURL: srv.URL + "/jmap",
			PrimaryAccounts: map[string]string{
				mailCapabilityURN: "account-123",
			},
			DefaultIdentity: map[string]string{
				mailCapabilityURN: "identity-789",
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	cfg := &config.RuntimeConfig{
		EmailJMAPEndpoint: srv.URL + "/jmap",
		EmailJMAPUser:     "user",
		EmailJMAPPass:     "pass",
	}

	p, err := providerFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected provider, got nil")
	}
	prov, ok := p.(Provider)
	if !ok {
		t.Fatalf("unexpected provider type: %#v", p)
	}

	if !sawWellKnown {
		t.Fatal("well-known endpoint not queried")
	}
	if prov.AccountID != "account-123" {
		t.Fatalf("unexpected account id: %s", prov.AccountID)
	}
	if prov.Identity != "identity-789" {
		t.Fatalf("unexpected identity id: %s", prov.Identity)
	}
	if prov.Endpoint != srv.URL+"/jmap" {
		t.Fatalf("unexpected endpoint: %s", prov.Endpoint)
	}
}
