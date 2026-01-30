package admin

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// GrantBulkCreateTask creates grants for multiple users or roles at once.
type GrantBulkCreateTask struct{ tasks.TaskString }

var grantBulkCreateTask = &GrantBulkCreateTask{TaskString: TaskGrantBulkCreate}

var _ tasks.Task = (*GrantBulkCreateTask)(nil)

func (GrantBulkCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	userIDs, err := parseIDs(r.PostForm["user_id"])
	if err != nil {
		return fmt.Errorf("parse user ids %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	roleIDs, err := parseIDs(r.PostForm["role_id"])
	if err != nil {
		return fmt.Errorf("parse role ids %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	anyone := r.PostFormValue("anyone") != ""
	if len(userIDs) == 0 && len(roleIDs) == 0 && !anyone {
		return fmt.Errorf("missing grant subject %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}

	sections := r.PostForm["section"]
	items := r.PostForm["item"]
	itemIDs := r.PostForm["item_id"]
	actionSets := r.PostForm["actions"]
	if len(sections) == 0 {
		return fmt.Errorf("missing grant sections %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	if len(items) != len(sections) || len(itemIDs) != len(sections) || len(actionSets) != len(sections) {
		return fmt.Errorf("grant rows mismatch %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}

	for i, section := range sections {
		item := items[i]
		actionSet := splitActions(actionSets[i])
		if section == "" || len(actionSet) == 0 {
			return fmt.Errorf("missing grant section or actions %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
		}
		itemID, err := parseNullableID(itemIDs[i])
		if err != nil {
			return fmt.Errorf("parse item id %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if def, ok := GrantActionMap[section+"|"+item]; ok && def.RequireItemID && !itemID.Valid {
			return fmt.Errorf("missing item id %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
		}

		for _, action := range actionSet {
			for _, userID := range userIDs {
				if _, err := queries.AdminCreateGrant(r.Context(), db.AdminCreateGrantParams{
					UserID:   sql.NullInt32{Int32: userID, Valid: true},
					RoleID:   sql.NullInt32{},
					Section:  section,
					Item:     sql.NullString{String: item, Valid: item != ""},
					RuleType: "allow",
					ItemID:   itemID,
					ItemRule: sql.NullString{},
					Action:   action,
					Extra:    sql.NullString{},
				}); err != nil {
					log.Printf("CreateGrant: %v", err)
					return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
				}
			}
			for _, roleID := range roleIDs {
				if _, err := queries.AdminCreateGrant(r.Context(), db.AdminCreateGrantParams{
					UserID:   sql.NullInt32{},
					RoleID:   sql.NullInt32{Int32: roleID, Valid: true},
					Section:  section,
					Item:     sql.NullString{String: item, Valid: item != ""},
					RuleType: "allow",
					ItemID:   itemID,
					ItemRule: sql.NullString{},
					Action:   action,
					Extra:    sql.NullString{},
				}); err != nil {
					log.Printf("CreateGrant: %v", err)
					return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
				}
			}
			if anyone {
				if _, err := queries.AdminCreateGrant(r.Context(), db.AdminCreateGrantParams{
					UserID:   sql.NullInt32{},
					RoleID:   sql.NullInt32{},
					Section:  section,
					Item:     sql.NullString{String: item, Valid: item != ""},
					RuleType: "allow",
					ItemID:   itemID,
					ItemRule: sql.NullString{},
					Action:   action,
					Extra:    sql.NullString{},
				}); err != nil {
					log.Printf("CreateGrant: %v", err)
					return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
				}
			}
		}
	}

	return handlers.RefreshDirectHandler{TargetURL: "/admin/grants"}
}

func parseIDs(values []string) ([]int32, error) {
	if len(values) == 0 {
		return nil, nil
	}
	ids := make([]int32, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		id, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		ids = append(ids, int32(id))
	}
	return ids, nil
}

func parseNullableID(value string) (sql.NullInt32, error) {
	if strings.TrimSpace(value) == "" {
		return sql.NullInt32{}, nil
	}
	id, err := strconv.Atoi(value)
	if err != nil {
		return sql.NullInt32{}, err
	}
	return sql.NullInt32{Int32: int32(id), Valid: true}, nil
}

func splitActions(value string) []string {
	parts := strings.Split(value, ",")
	actions := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			actions = append(actions, part)
		}
	}
	return actions
}
