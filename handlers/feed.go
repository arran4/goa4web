package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/gorilla/mux"
)

// VerifyFeedRequest verifies the signature of a signed feed request.
// It returns the user who signed the request, or an error.
func VerifyFeedRequest(r *http.Request, basePath string) (*db.User, error) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		return nil, fmt.Errorf("username not found in vars")
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if cd.FeedSignKey == "" {
		return nil, fmt.Errorf("feed signing key not configured")
	}

	sig := r.URL.Query().Get("sig")
	if sig == "" {
		return nil, fmt.Errorf("missing signature")
	}

	// Build data to verify
	data := "feed:" + username + ":" + basePath

	// Verify signature (feeds use WithOutNonce by default)
	if err := sign.Verify(data, sig, cd.FeedSignKey, sign.WithOutNonce()); err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	queries := cd.Queries()
	u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		return nil, err
	}
	return &db.User{Idusers: u.Idusers, Username: u.Username}, nil
}
