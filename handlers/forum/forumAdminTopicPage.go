package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminTopicPage shows information about a single forum topic.
func AdminTopicPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	topic, err := cd.ForumTopicByID(int32(tid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Topic not found"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum Topic %d", tid)
	data := struct {
		Topic *db.GetForumTopicByIdForUserRow
	}{
		Topic: topic,
	}
	handlers.TemplateHandler(w, r, "forum/adminTopicPage.gohtml", data)
}

// AdminTopicEditFormPage shows the edit form for a forum topic.
func AdminTopicEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	topic, err := cd.ForumTopicByID(int32(tid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Topic not found"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	categories, err := cd.ForumCategories()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	roles, err := cd.AllRoles()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Edit Forum Topic %d", tid)
	data := struct {
		Topic       *db.GetForumTopicByIdForUserRow
		Categories  []*db.Forumcategory
		Roles       []*db.Role
		Restriction interface{}
	}{
		Topic:      topic,
		Categories: categories,
		Roles:      roles,
	}
	handlers.TemplateHandler(w, r, "forum/adminTopicEditPage.gohtml", data)
}
