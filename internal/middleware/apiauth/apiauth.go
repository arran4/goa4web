package apiauth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func writeAPIError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": message})
}

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
			writeAPIError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		hash := sha256.Sum256([]byte(token))
		tokenHash := hex.EncodeToString(hash[:])

		apiKey, err := cd.Queries().GetAPIKeyByHash(r.Context(), tokenHash)
		if err != nil {
			// Token not found or invalid
			writeAPIError(w, http.StatusUnauthorized, "unauthorized")
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

		// Create a fully hydrated CoreData for the user rather than mutating the anonymous one
		userCD := cd.ForUser(apiKey.UsersIdusers)

		// Preserve request event if one exists from prior middleware (e.g., TaskEventMiddleware)
		if evt := cd.Event(); evt != nil {
			evt.UserID = apiKey.UsersIdusers
			userCD.SetEvent(evt)
		}

		// Store scopes in context for specific handlers to check
		ctx := context.WithValue(r.Context(), consts.KeyCoreData, userCD)
		ctx = context.WithValue(ctx, consts.KeyAPIScopes, scopeMap)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireScope is a middleware that enforces that the API key has a specific scope.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scopes, ok := r.Context().Value(consts.KeyAPIScopes).(map[string]bool)
			if !ok {
				writeAPIError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if !scopes[scope] {
				writeAPIError(w, http.StatusForbidden, "required API scope is missing")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
