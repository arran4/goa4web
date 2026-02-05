package faq

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
	"github.com/gorilla/mux"
)

// AddCategoryGrantTask creates a new grant for an FAQ category.
type AddCategoryGrantTask struct{ tasks.TaskString }

var addCategoryGrantTask = &AddCategoryGrantTask{TaskString: TaskAddCategoryGrant}

var _ tasks.Task = (*AddCategoryGrantTask)(nil)

func (AddCategoryGrantTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskAddCategoryGrant)(r, m)
}

func (AddCategoryGrantTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["id"])
	if err != nil {
		categoryID, err = strconv.Atoi(r.PostFormValue("category_id"))
		if err != nil {
			return fmt.Errorf("category id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")

	if username == "" && role == "" {
		return fmt.Errorf("username or role required %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("username or role required")))
	}

	actions := r.Form["action"]
	if len(actions) == 0 {
		val := r.PostFormValue("action")
		if val != "" {
			actions = []string{val}
		} else {
			actions = []string{"see"}
		}
	}
	var uid sql.NullInt32
	if username != "" {
		u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("SystemGetUserByUsername: %v", err)
			return fmt.Errorf("get user by username %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		uid = sql.NullInt32{Int32: u.Idusers, Valid: true}
	}
	var rid sql.NullInt32
	if role != "" {
		roles, err := queries.AdminListRoles(r.Context())
		if err != nil {
			log.Printf("ListRoles: %v", err)
			return fmt.Errorf("list roles %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		for _, ro := range roles {
			if ro.Name == role {
				rid = sql.NullInt32{Int32: ro.ID, Valid: true}
				break
			}
		}
	}
	for _, action := range actions {
		if action == "" {
			continue
		}
		if _, err = queries.AdminCreateGrant(r.Context(), db.AdminCreateGrantParams{
			UserID:   uid,
			RoleID:   rid,
			Section:  "faq",
			Item:     sql.NullString{String: "category", Valid: true},
			RuleType: "allow",
			ItemID:   sql.NullInt32{Int32: int32(categoryID), Valid: true},
			ItemRule: sql.NullString{},
			Action:   action,
			Extra:    sql.NullString{},
		}); err != nil {
			log.Printf("CreateGrant: %v", err)
			return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return nil
}

// RemoveCategoryGrantTask removes a grant from an FAQ category.
type RemoveCategoryGrantTask struct{ tasks.TaskString }

var removeCategoryGrantTask = &RemoveCategoryGrantTask{TaskString: TaskRemoveCategoryGrant}

var _ tasks.Task = (*RemoveCategoryGrantTask)(nil)

func (RemoveCategoryGrantTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskRemoveCategoryGrant)(r, m)
}

func (RemoveCategoryGrantTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	grantID, err := strconv.Atoi(r.PostFormValue("grant_id"))
	if err != nil {
		return fmt.Errorf("grant id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	vars := mux.Vars(r)
	categoryID, _ := strconv.Atoi(vars["id"])
	if categoryID == 0 {
		categoryID, _ = strconv.Atoi(r.PostFormValue("category_id"))
	}

	if err := cd.Queries().AdminDeleteGrant(r.Context(), int32(grantID)); err != nil {
		return fmt.Errorf("delete grant fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
