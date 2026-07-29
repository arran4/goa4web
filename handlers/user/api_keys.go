package user

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/segmentio/ksuid"
)

func DownloadSwagger(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Content-Disposition", "attachment; filename=\"swagger.yaml\"")

	// Pass the BaseURL to the template to populate the server URL correctly
	tasks.Template("user/swagger.gohtml").Handle(w, r, struct{
		BaseURL string
		Version string
	}{
		BaseURL: strings.TrimRight(cd.Config.BaseURL, "/"),
		Version: goa4web.Version,
	})
}

func ListAPIKeysPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	keys, err := cd.Queries().ListAPIKeysByUser(r.Context(), cd.UserID)
	if err != nil && err != sql.ErrNoRows {
		handlers.RenderErrorPage(w, r, fmt.Errorf("failed to list api keys: %w", err))
		return
	}

	tasks.Template("user/apiKeysPage.gohtml").Handle(w, r, struct {
		Keys []*db.ApiKey
	}{
		Keys: keys,
	})
}

type CreateAPIKeyTask struct{ tasks.TaskString }

var createAPIKeyTask = &CreateAPIKeyTask{TaskString: "create-api-key"}

func (CreateAPIKeyTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return fmt.Errorf("name is required")
	}

	r.ParseForm()
	scopes := r.Form["scopes"]
	if len(scopes) == 0 {
		return fmt.Errorf("at least one scope is required")
	}

	validScopes := map[string]bool{
		"private_forum:read":  true,
		"private_forum:write": true,
		"images:read":         true,
		"images:write":        true,
	}

	var validatedScopes []string
	for _, s := range scopes {
		if validScopes[s] {
			validatedScopes = append(validatedScopes, s)
		}
	}

	if len(validatedScopes) == 0 {
		return fmt.Errorf("invalid scopes")
	}

	// Generate random API key using ksuid + random parts
	rawKey := fmt.Sprintf("goa4web_%s", ksuid.New().String())

	hash := sha256.Sum256([]byte(rawKey))
	tokenHash := hex.EncodeToString(hash[:])

	// TTL
	var expiresAt sql.NullTime
	ttlStr := r.FormValue("ttl_days")
	if ttlStr != "" {
		var days int
		fmt.Sscanf(ttlStr, "%d", &days)
		if days > 0 {
			expiresAt.Time = time.Now().AddDate(0, 0, days)
			expiresAt.Valid = true
		}
	}

	_, err := cd.Queries().CreateAPIKey(r.Context(), db.CreateAPIKeyParams{
		UsersIdusers: cd.UserID,
		ApiKey:       tokenHash,
		Name:         name,
		Scopes:       strings.Join(validatedScopes, ","),
		ExpiresAt:    expiresAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create api key: %w", err)
	}

	cd.SetCurrentNotice(fmt.Sprintf("API Key created! Save it now, you won't be able to see it again: %s", rawKey))

	return "/usr/api-keys"
}

type RevokeAPIKeyTask struct{ tasks.TaskString }

var revokeAPIKeyTask = &RevokeAPIKeyTask{TaskString: "revoke-api-key"}

func (RevokeAPIKeyTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	var keyID int32
	if _, err := fmt.Sscanf(r.FormValue("key_id"), "%d", &keyID); err != nil || keyID <= 0 {
		return fmt.Errorf("invalid key id")
	}

	err := cd.Queries().RevokeAPIKey(r.Context(), db.RevokeAPIKeyParams{
		ID:           keyID,
		UsersIdusers: cd.UserID,
	})

	if err != nil {
		return fmt.Errorf("failed to revoke api key: %w", err)
	}

	cd.SetCurrentNotice("API Key revoked successfully.")
	return "/usr/api-keys"
}
