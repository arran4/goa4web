package csrf

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"filippo.io/csrf/gorilla"

	"github.com/arran4/goa4web/core"
)

type contextKey string

const (
	// contextTokenKey stores the CSRF token in the request context.
	contextTokenKey contextKey = "csrfToken"
	// sessionTokenKey stores the CSRF token alongside the user's session data.
	sessionTokenKey = "CSRFToken"
	// sessionUserKey tracks the user ID associated with the CSRF token.
	sessionUserKey = "CSRFAuthUID"
	// formFieldName is the expected CSRF form field name.
	formFieldName = "gorilla.csrf.Token"
)

type lazyCSRF struct {
	w     http.ResponseWriter
	r     *http.Request
	token string
}

func (l *lazyCSRF) getToken(currentW http.ResponseWriter, currentR *http.Request) string {
	if l.token != "" {
		return l.token
	}

	session, err := core.GetSession(currentR)
	if err != nil {
		core.SessionErrorRedirect(currentW, currentR, err)
		return ""
	}
	currentUID := readUID(session.Values["UID"])
	tokenUID := readUID(session.Values[sessionUserKey])
	token, _ := session.Values[sessionTokenKey].(string)

	if token == "" || currentUID != tokenUID {
		token, err = newToken()
		if err != nil {
			log.Printf("generate csrf token: %v", err)
			http.Error(currentW, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return ""
		}
		session.Values[sessionTokenKey] = token
		session.Values[sessionUserKey] = currentUID
		if err := session.Save(currentR, currentW); err != nil {
			log.Printf("save csrf token: %v", err)
			http.Error(currentW, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return ""
		}
		// A new session/token was generated and saved to the browser.
		// This must not be publicly cacheable.
		core.DisableCaching(currentW)
	}

	l.token = token
	return token
}

// NewCSRFMiddleware returns middleware enforcing CSRF protection using the
// provided session secret and HTTP configuration. It also issues per-session
// CSRF tokens that rotate when the authenticated user changes.
func NewCSRFMiddleware(secret string, hostname string, version string) func(http.Handler) http.Handler {
	key := sha256.Sum256([]byte(secret))
	origins := []string{}
	if u, err := url.Parse(hostname); err == nil && u.Host != "" {
		origins = append(origins, u.Host)
	}
	//nolint:staticcheck // gorilla/csrf deprecation
	protect := csrf.Protect(key[:], csrf.Secure(version != "dev"), csrf.TrustedOrigins(origins))
	return func(next http.Handler) http.Handler {
		validatedNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requiresToken(r.Method) && !validateRequestToken(r) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
		protected := protect(validatedNext)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lazy := &lazyCSRF{w: w, r: r}
			ctx := context.WithValue(r.Context(), contextTokenKey, lazy)
			protected.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Token returns the request-specific CSRF token.
func Token(r *http.Request) string {
	val := r.Context().Value(contextTokenKey)
	if lazy, ok := val.(*lazyCSRF); ok {
		// Use the currently active response writer from the context if available
		// (e.g., a buffer provided by TemplateHandler) to ensure redirects or errors
		// triggered by lazy generation do not bypass the buffering bounds.
		cw := core.GetCurrentResponseWriter(r)
		if cw == nil {
			cw = lazy.w
		}
		return lazy.getToken(cw, r)
	}
	if token, ok := val.(string); ok {
		return token
	}
	return ""
}

// TemplateField returns the HTML hidden input tag containing the CSRF token.
func TemplateField(r *http.Request) template.HTML {
	token := Token(r)
	if token == "" {
		return template.HTML("")
	}
	return template.HTML(fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, formFieldName, template.HTMLEscapeString(token)))
}

func validateRequestToken(r *http.Request) bool {
	token := Token(r)
	if token == "" {
		return false
	}
	if header := r.Header.Get("X-CSRF-Token"); header != "" {
		return subtleCompare(header, token)
	}
	if err := r.ParseForm(); err != nil {
		return false
	}
	return subtleCompare(r.PostFormValue(formFieldName), token)
}

func subtleCompare(provided string, expected string) bool {
	if provided == "" || expected == "" {
		return false
	}
	if len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func requiresToken(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	default:
		return true
	}
}

func newToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(token), nil
}

func readUID(uid any) int32 {
	switch v := uid.(type) {
	case int:
		return int32(v)
	case int32:
		return v
	case int64:
		return int32(v)
	case float64:
		return int32(v)
	default:
		return 0
	}
}
