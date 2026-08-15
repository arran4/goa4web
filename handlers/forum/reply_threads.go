package forum

import (
	"fmt"
	"net/http"
	"github.com/arran4/goa4web/core/common"
	"strconv"


	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"

	"database/sql"
	"github.com/gorilla/mux"
)

const ReplyThreadsPageTmpl tasks.Template = "domains/forum/replythreadsPage.gohtml"




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
		Thread       *db.GetThreadLastPosterAndPermsForUserRow
		ReplyThreads []*db.GetReplyThreadsForThreadRow
	}

	uid, _ := cd.GetSession().Values["UID"].(int32)
	queries := cd.Queries()

	topic, err := queries.GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{
		ViewerID:      uid,
		Idforumtopic:  int32(topicId),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("topic not found"))
		return
	}

	thread, err := queries.GetThreadLastPosterAndPermsForUser(r.Context(), db.GetThreadLastPosterAndPermsForUserParams{
		ViewerID:      uid,
		ThreadID: int32(threadId),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("thread not found"))
		return
	}

	replyThreads, err := queries.GetReplyThreadsForThread(r.Context(), sql.NullInt32{Int32: int32(threadId), Valid: threadId != 0})
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
