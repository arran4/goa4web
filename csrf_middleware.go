package main

import (
	"crypto/sha256"
	"net/url"

	"github.com/gorilla/csrf"
)

// newCSRFMiddleware returns a wrapper enforcing CSRF protection using the
// provided session secret and HTTP configuration.
func newCSRFMiddleware(secret string, hostname string, version string) routerWrapper {
	key := sha256.Sum256([]byte(secret))
	origins := []string{}
	if u, err := url.Parse(hostname); err == nil && u.Host != "" {
		origins = append(origins, u.Host)
	}
	return routerWrapperFunc(csrf.Protect(key[:], csrf.Secure(version != "dev"), csrf.TrustedOrigins(origins)))
}
