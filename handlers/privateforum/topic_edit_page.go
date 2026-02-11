package privateforum

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

const TopicEditPageTmpl tasks.Template = "forum/topicEditPage.gohtml"

func TopicEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
		return
	}

	if !cd.HasGrant("privateforum", "topic", "edit", int32(topicID)) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("permission denied"))
		return
	}

	topic, err := cd.Queries().GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{
		ViewerID:      cd.UserID,
		Idforumtopic:  int32(topicID),
		ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		log.Printf("GetForumTopicByIdForUser: %v", err)
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
		return
	}

	data := struct {
		Topic    *db.GetForumTopicByIdForUserRow
		BasePath string
	}{
		Topic:    topic,
		BasePath: "/private",
	}
	if cd.ForumBasePath != "" {
		data.BasePath = cd.ForumBasePath
	}

	TopicEditPageTmpl.Handle(w, r, data)
}

func TopicEditSubmit(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
		return
	}

	if !cd.HasGrant("privateforum", "topic", "edit", int32(topicID)) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("permission denied"))
		return
	}

	if err := r.ParseForm(); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	if title == "" {
		cd.SetCurrentError("Title cannot be empty")
		TopicEditPage(w, r)
		return
	}

	// Load existing topic to get category and lang
	topic, err := cd.Queries().GetForumTopicById(r.Context(), int32(topicID))
	if err != nil {
		log.Printf("GetForumTopicById: %v", err)
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
		return
	}

	err = cd.Queries().AdminUpdateForumTopic(r.Context(), db.AdminUpdateForumTopicParams{
		Title:                        sql.NullString{String: title, Valid: true},
		Description:                  sql.NullString{String: description, Valid: true},
		ForumcategoryIdforumcategory: topic.ForumcategoryIdforumcategory,
		TopicLanguageID:              topic.LanguageID,
		Idforumtopic:                 int32(topicID),
	})

	if err != nil {
		log.Printf("AdminUpdateForumTopic: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	basePath := "/private"
	if cd.ForumBasePath != "" {
		basePath = cd.ForumBasePath
	}

	http.Redirect(w, r, fmt.Sprintf("%s/topic/%d", basePath, topicID), http.StatusSeeOther)
}
