cat << 'INNER_EOF' > cmd/goa4web/scenario_serve_csrf_sqlite_test.go
//go:build sqlite
// +build sqlite

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
    "os"
    "fmt"
)

func TestScenarioServeCmd_CSRFCachingGuarantees(t *testing.T) {
	ctx := context.Background()

    os.Setenv("GOA4WEB_CSRF_ENABLED", "true")
    os.Setenv("GOA4WEB_SESSION_SECRET", "testsecret")
    defer os.Unsetenv("GOA4WEB_CSRF_ENABLED")
    defer os.Unsetenv("GOA4WEB_SESSION_SECRET")

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "../../testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "../../testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	cmd, err := parseScenarioServeCmd(parent, []string{"../../testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}

	srv, _, cleanup, err := cmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("scenarioServeCmd Bootstrap failed: %v", err)
	}
	defer cleanup()

	router := srv.Router

    t.Run("anonymous_read_only", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "http://localhost/faq", nil)
        rr := httptest.NewRecorder()

        router.ServeHTTP(rr, req)

        res := rr.Result()
        if res.StatusCode != http.StatusOK {
            t.Fatalf("Expected 200 OK for /faq, got %d", res.StatusCode)
        }

        for _, c := range res.Header.Values("Set-Cookie") {
            if strings.HasPrefix(c, "my-session=") || strings.HasPrefix(c, "goa4web_session=") {
                t.Errorf("Anonymous read-only GET incorrectly emitted application session cookie: %s", c)
            }
        }

        cc := res.Header.Get("Cache-Control")
        if cc != "public, max-age=3600" {
            t.Errorf("Expected Cache-Control to be public, max-age=3600; got %q", cc)
        }
    })

    t.Run("anonymous_csrf_form", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "http://localhost/login", nil)
        rr := httptest.NewRecorder()

        router.ServeHTTP(rr, req)

        res := rr.Result()
        if res.StatusCode != http.StatusOK {
            t.Fatalf("Expected 200 OK for /login, got %d", res.StatusCode)
        }

        hasAppCookie := false
        for _, c := range res.Header.Values("Set-Cookie") {
            fmt.Printf("Set-Cookie Header: %s\n", c)
            if strings.HasPrefix(c, "my-session=") || strings.HasPrefix(c, "goa4web_session=") {
                hasAppCookie = true
            }
        }
        if !hasAppCookie {
            t.Errorf("Expected application session cookie for /login because it contains a CSRF form")
        }

        cc := res.Header.Get("Cache-Control")
        expectedCC := "no-cache, no-store, must-revalidate"
        if cc != expectedCC {
            t.Errorf("Expected Cache-Control %q, got %q", expectedCC, cc)
        }
        cfcc := res.Header.Get("Cloudflare-CDN-Cache-Control")
        expectedCFCC := "no-store"
        if cfcc != expectedCFCC {
            t.Errorf("Expected Cloudflare-CDN-Cache-Control %q, got %q", expectedCFCC, cfcc)
        }
    })
}
INNER_EOF
