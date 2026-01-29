package forum

import (
	"github.com/arran4/goa4web/handlers/forum/forumcommon"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// TopicGrantCreateTask creates a new grant for a forum topic.
type TopicGrantCreateTask struct{ tasks.TaskString }

var topicGrantCreateTask = &TopicGrantCreateTask{TaskString: forumcommon.TaskTopicGrantCreate}

var _ tasks.Task = (*TopicGrantCreateTask)(nil)

func (TopicGrantCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		return fmt.Errorf("topic id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")
	actions := r.Form["action"]
	if len(actions) == 0 {
		actions = []string{"see"}
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
			action = "see"
		}
		if _, err = cd.GrantForumTopic(int32(topicID), uid, rid, action); err != nil {
			log.Printf("CreateGrant: %v", err)
			return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/forum/topics/topic/%d/grants", topicID)}
}
