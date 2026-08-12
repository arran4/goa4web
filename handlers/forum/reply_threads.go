package forum

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

var ReplyThreadsPageTmpl = handlers.Template{
	Path:   "domains/forum/replythreadsPage",
	Target: "core/templates/site/domains/forum/replythreadsPage.gohtml",
}.Must()

func ReplyThreadsPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	topicId, err := strconv.Atoi(vars["topic"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid topic ID"))
		return
	}

	threadId, err := strconv.Atoi(vars["thread"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid thread ID"))
		return
	}

	type Data struct {
		Topic        *db.GetForumTopicByIdForUserRow
		Thread       *db.GetForumThreadByIDForUserRow
		ReplyThreads []*db.GetReplyThreadsForThreadRow
	}

	uid, _ := cd.GetSession().Values["UID"].(int32)
	queries := cd.Queries()

	topic, err := queries.GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{
		ViewerID:      uid,
		Idforumtopic:  int32(topicId),
		ViewerMatchID: common.SqlNullInt32ZeroUnset(uid),
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("topic not found"))
		return
	}

	thread, err := queries.GetForumThreadByIDForUser(r.Context(), db.GetForumThreadByIDForUserParams{
		ViewerID:      uid,
		Idforumthread: int32(threadId),
		ViewerMatchID: common.SqlNullInt32ZeroUnset(uid),
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("thread not found"))
		return
	}

	replyThreads, err := queries.GetReplyThreadsForThread(r.Context(), common.SqlNullInt32ZeroUnset(int32(threadId)))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("error fetching reply threads"))
		return
	}

	cd.PageTitle = "Reply Threads - " + topic.Title.String
	data := Data{
		Topic:        topic,
		Thread:       thread,
		ReplyThreads: replyThreads,
	}

	if err := ReplyThreadsPageTmpl.Handle(w, r, data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
