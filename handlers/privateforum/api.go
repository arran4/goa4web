package privateforum

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type UserExistsResponse struct {
	Exists bool `json:"exists"`
}

var ErrUserNotFound = errors.New("user not found")

func UserExistsAPI(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	username := r.FormValue("username")
	exists, err := userExists(r.Context(), queries, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UserExistsResponse{Exists: false})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserExistsResponse{Exists: exists})
}

func userExists(ctx context.Context, queries db.Querier, username string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("username is required")
	}
	_, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: username, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrUserNotFound
		}
		return false, err
	}
	return true, nil
}
