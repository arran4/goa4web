package apiauth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// APIKeyAuthMiddleware checks for a Bearer token, validates it, and sets the user identity.
func APIKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			next.ServeHTTP(w, r) // No token, proceed as anonymous
			return
		}

		token := strings.TrimSpace(authHeader[7:])
		if token == "" {
			handlers.RenderErrorPage(w, r, handlers.ErrUnauthorized)
			return
		}

		hash := sha256.Sum256([]byte(token))
		tokenHash := hex.EncodeToString(hash[:])

		apiKey, err := cd.Queries().GetAPIKeyByHash(r.Context(), tokenHash)
		if err != nil {
			// Token not found or invalid
			handlers.RenderErrorPage(w, r, handlers.ErrUnauthorized)
			return
		}

		// Update last used
		_ = cd.Queries().UpdateAPIKeyLastUsed(r.Context(), apiKey.ID)

		// Parse scopes
		scopes := strings.Split(apiKey.Scopes, ",")
		scopeMap := make(map[string]bool)
		for _, s := range scopes {
			scopeMap[strings.TrimSpace(s)] = true
		}

		// Set UserID in CoreData to act as the user
		cd.UserID = apiKey.UsersIdusers

		// Store scopes in context for specific handlers to check
		ctx := context.WithValue(r.Context(), consts.KeyAPIScopes, scopeMap)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireScope is a middleware that enforces that the API key has a specific scope.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scopes, ok := r.Context().Value(consts.KeyAPIScopes).(map[string]bool)
			if !ok || !scopes[scope] {
				handlers.RenderErrorPage(w, r, handlers.ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
