package csrf

import (
	"crypto/sha256"
	"net/http"
	"net/url"
	"os"
	"strings"

	config "github.com/arran4/goa4web/config"
	"github.com/gorilla/csrf"
)

// CSRFEnabled reports if CSRF protection should be active.
func CSRFEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvCSRFEnabled))
	if v == "" {
		return true
	}
	switch v {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

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
