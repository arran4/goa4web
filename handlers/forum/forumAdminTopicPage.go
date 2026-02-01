package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"strconv"
	"strings"

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
	topic, err := cd.Queries().GetForumTopicById(r.Context(), int32(tid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("AdminTopicPage: Topic %d not found", tid)
			w.WriteHeader(http.StatusNotFound)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Topic not found"))
			return
		}
		log.Printf("AdminTopicPage: Error fetching topic %d: %v", tid, err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum Topic %d", tid)
	// Check if "anyone" (public) has access to this topic.
	// This logic depends on how "anyone" access is modeled in grants.
	// Assuming a grant exists for role "Anyone" or similar logic.
	// For now, we'll check if there's any grant for the topic that applies to public.
	// Since we don't have a direct "IsPublic" helper, we'll fetch grants and check in memory or use a new query.
	// Using AdminListForumTopicGrantsByTopicID.
	grants, err := cd.Queries().AdminListForumTopicGrantsByTopicID(r.Context(), sql.NullInt32{Int32: int32(tid), Valid: true})
	anyoneHasAccess := false
	if err == nil {
		for _, g := range grants {
			if g.RoleName.Valid && strings.EqualFold(g.RoleName.String, "Anyone") && g.Action == "see" {
				anyoneHasAccess = true
				break
			}
		}
	}

	var participants []*db.AdminListPrivateTopicParticipantsByTopicIDRow
	if topic.Handler == "private" {
		participants, err = cd.Queries().AdminListPrivateTopicParticipantsByTopicID(r.Context(), sql.NullInt32{Int32: int32(tid), Valid: true})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error fetching participants"))
			return
		}
	}

	var labels []string
	if pub, _, err := cd.TopicPublicLabels(topic.Idforumtopic); err == nil {
		labels = pub
	} else {
		log.Printf("list public labels: %v", err)
	}

	data := struct {
		Topic           *db.Forumtopic
		AnyoneHasAccess bool
		Participants    []*db.AdminListPrivateTopicParticipantsByTopicIDRow
		Labels          []string
	}{
		Topic:           topic,
		AnyoneHasAccess: anyoneHasAccess,
		Participants:    participants,
		Labels:          labels,
	}
	ForumAdminTopicPageTmpl.Handle(w, r, data)
}

const ForumAdminTopicPageTmpl tasks.Template = "forum/adminTopicPage.gohtml"

// AdminTopicEditFormPage shows the edit form for a forum topic.
func AdminTopicEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	topic, err := cd.Queries().GetForumTopicById(r.Context(), int32(tid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("AdminTopicEditFormPage: Topic %d not found", tid)
			w.WriteHeader(http.StatusNotFound)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Topic not found"))
			return
		}
		log.Printf("AdminTopicEditFormPage: Error fetching topic %d: %v", tid, err)
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
		Topic       *db.Forumtopic
		Categories  []*db.Forumcategory
		Roles       []*db.Role
		Restriction interface{}
	}{
		Topic:      topic,
		Categories: categories,
		Roles:      roles,
	}
	ForumAdminTopicEditPageTmpl.Handle(w, r, data)
}

const ForumAdminTopicEditPageTmpl tasks.Template = "forum/adminTopicEditPage.gohtml"
