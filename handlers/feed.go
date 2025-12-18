package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/feedsign"
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

	if cd.FeedSigner == nil {
		return nil, fmt.Errorf("feed signer not configured")
	}

	ts := r.URL.Query().Get("ts")
	sig := r.URL.Query().Get("sig")
	originalQuery := feedsign.StripSignatureParams(r.URL.Query())

	if !cd.FeedSigner.Verify(basePath, originalQuery, username, ts, sig) {
		return nil, fmt.Errorf("invalid signature")
	}

	queries := cd.Queries()
	u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		return nil, err
	}
	return &db.User{Idusers: u.Idusers, Username: u.Username}, nil
}
