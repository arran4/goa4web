package forum

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminTopicPage shows a dashboard for a single forum topic.
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
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Topic not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum Topic %d", tid)
	data := struct {
		Topic *db.Forumtopic
	}{
		Topic: topic,
	}
	handlers.TemplateHandler(w, r, "adminTopicPage.gohtml", data)
}

// AdminTopicThreadsPage shows all threads for a single forum topic.
func AdminTopicThreadsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	topic, err := cd.Queries().GetForumTopicById(r.Context(), int32(tid))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Topic not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum Topic %d Threads", tid)
	data := struct {
		Topic *db.Forumtopic
	}{
		Topic: topic,
	}
	handlers.TemplateHandler(w, r, "adminTopicThreadsPage.gohtml", data)
}
