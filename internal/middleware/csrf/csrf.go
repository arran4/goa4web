package csrf

import (
	"crypto/sha256"
	"html/template"
	"net/http"
	"net/url"

	"filippo.io/csrf/gorilla"
)

// NewCSRFMiddleware returns middleware enforcing CSRF protection using the
// provided session secret and HTTP configuration.
func NewCSRFMiddleware(secret string, hostname string, version string) func(http.Handler) http.Handler {
	key := sha256.Sum256([]byte(secret))
	origins := []string{}
	if u, err := url.Parse(hostname); err == nil && u.Host != "" {
		origins = append(origins, u.Host)
	}
	return csrf.Protect(key[:], csrf.Secure(version != "dev"), csrf.TrustedOrigins(origins))
}

// Token returns the request-specific CSRF token.
func Token(r *http.Request) string {
	return csrf.Token(r)
}

// TemplateField returns the HTML hidden input tag containing the CSRF token.
func TemplateField(r *http.Request) template.HTML {
	return csrf.TemplateField(r)
}
