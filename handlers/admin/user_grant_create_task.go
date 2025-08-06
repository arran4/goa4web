package admin

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserGrantCreateTask creates a new grant for a user.
type UserGrantCreateTask struct{ tasks.TaskString }

var userGrantCreateTask = &UserGrantCreateTask{TaskString: TaskRoleGrantCreate}

var _ tasks.Task = (*UserGrantCreateTask)(nil)

func (UserGrantCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	userID := cd.CurrentProfileUserID()
	if userID == 0 {
		return fmt.Errorf("user id parse fail %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	section := r.PostFormValue("section")
	item := r.PostFormValue("item")
	action := r.PostFormValue("action")
	itemIDStr := r.PostFormValue("item_id")
	var itemID sql.NullInt32
	if itemIDStr != "" {
		id, err := strconv.Atoi(itemIDStr)
		if err != nil {
			return fmt.Errorf("item id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		itemID = sql.NullInt32{Int32: int32(id), Valid: true}
	}
	if section == "" || action == "" {
		return fmt.Errorf("missing section or action %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	if def, ok := GrantActionMap[section+"|"+item]; ok && def.RequireItemID && !itemID.Valid {
		return fmt.Errorf("missing item id %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
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
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d/grants", userID)}
}
