package forum

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

// TopicGrantCreateTask creates a new grant for a forum topic.
type TopicGrantCreateTask struct{ tasks.TaskString }

var topicGrantCreateTask = &TopicGrantCreateTask{TaskString: TaskTopicGrantCreate}

var _ tasks.Task = (*TopicGrantCreateTask)(nil)

func (TopicGrantCreateTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")
	actions := r.Form["action"]
	if len(actions) == 0 {
		actions = []string{"see"}
	}
	var uid sql.NullInt32
	if username != "" {
		u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("GetUserByUsername: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		uid = sql.NullInt32{Int32: u.Idusers, Valid: true}
	}
	var rid sql.NullInt32
	if role != "" {
		roles, err := queries.ListRoles(r.Context())
		if err != nil {
			log.Printf("ListRoles: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
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
		if _, err = queries.CreateGrant(r.Context(), db.CreateGrantParams{
			UserID:   uid,
			RoleID:   rid,
			Section:  "forum",
			Item:     sql.NullString{String: "topic", Valid: true},
			RuleType: "allow",
			ItemID:   sql.NullInt32{Int32: int32(topicID), Valid: true},
			ItemRule: sql.NullString{},
			Action:   action,
			Extra:    sql.NullString{},
		}); err != nil {
			log.Printf("CreateGrant: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

// TopicGrantDeleteTask removes a grant from a forum topic.
type TopicGrantDeleteTask struct{ tasks.TaskString }

var topicGrantDeleteTask = &TopicGrantDeleteTask{TaskString: TaskTopicGrantDelete}

var _ tasks.Task = (*TopicGrantDeleteTask)(nil)

func (TopicGrantDeleteTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	grantID, err := strconv.Atoi(r.PostFormValue("grantid"))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := queries.DeleteGrant(r.Context(), int32(grantID)); err != nil {
		log.Printf("DeleteGrant: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}
