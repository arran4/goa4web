package privateforum

import (
	"database/sql"
	"errors"
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
	"github.com/gorilla/mux"
)

const TopicEditPageTmpl tasks.Template = "domains/forum/topicEditPage.gohtml"

func TopicEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
		return
	}

	if !cd.HasGrant("privateforum", "topic", "edit", int32(topicID)) &&
		!cd.HasGrant("privateforum", "topic", "edit", 0) {
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

	_ = TopicEditPageTmpl.Handle(w, r, data)
}

func TopicEditSubmit(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
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

	err = cd.EditPrivateTopic(r.Context(), common.EditPrivateTopicParams{
		ActorID:     cd.UserID,
		TopicID:     int32(topicID),
		Title:       title,
		Description: description,
	})
	if err != nil {
		if errors.Is(err, common.ForumResourceNotFoundError{Resource: "topic"}) || strings.Contains(err.Error(), "not found") {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}
		var forbidden common.ForumOperationForbiddenError
		if errors.As(err, &forbidden) {
			handlers.RenderErrorPage(w, r, fmt.Errorf("permission denied"))
			return
		}
		log.Printf("EditPrivateTopic: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	basePath := "/private"
	if cd.ForumBasePath != "" {
		basePath = cd.ForumBasePath
	}

	http.Redirect(w, r, fmt.Sprintf("%s/topic/%d", basePath, topicID), http.StatusSeeOther)
}
