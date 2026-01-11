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

// NewCSRFMiddleware returns middleware enforcing CSRF protection using the
// provided session secret and HTTP configuration. It also issues per-session
// CSRF tokens that rotate when the authenticated user changes.
func NewCSRFMiddleware(secret string, hostname string, version string) func(http.Handler) http.Handler {
	key := sha256.Sum256([]byte(secret))
	origins := []string{}
	if u, err := url.Parse(hostname); err == nil && u.Host != "" {
		origins = append(origins, u.Host)
	}
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
			withToken, ok := attachToken(w, r)
			if !ok {
				return
			}
			protected.ServeHTTP(w, withToken)
		})
	}
}

func attachToken(w http.ResponseWriter, r *http.Request) (*http.Request, bool) {
	session, err := core.GetSession(r)
	if err != nil {
		core.SessionErrorRedirect(w, r, err)
		return nil, false
	}
	currentUID := readUID(session.Values["UID"])
	tokenUID := readUID(session.Values[sessionUserKey])
	token, _ := session.Values[sessionTokenKey].(string)

	if token == "" || currentUID != tokenUID {
		token, err = newToken()
		if err != nil {
			log.Printf("generate csrf token: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, false
		}
		session.Values[sessionTokenKey] = token
		session.Values[sessionUserKey] = currentUID
		if err := session.Save(r, w); err != nil {
			log.Printf("save csrf token: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, false
		}
	}

	ctx := context.WithValue(r.Context(), contextTokenKey, token)
	return r.WithContext(ctx), true
}

// Token returns the request-specific CSRF token.
func Token(r *http.Request) string {
	if token, ok := r.Context().Value(contextTokenKey).(string); ok {
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
