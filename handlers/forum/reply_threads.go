package forum

import (
	"fmt"
	"net/http"
	"strconv"
	"database/sql"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

const ReplyThreadsPageTmpl tasks.Template = "domains/forum/repliesPage.gohtml"

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

	type threadWithLabels struct {
		db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow
		Labels []templates.TopicLabel
	}

	type Data struct {
		Topic        *db.GetForumTopicByIdForUserRow
		Thread       *db.GetThreadLastPosterAndPermsForUserRow
		ThreadsByComment map[int32][]*threadWithLabels
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
		ThreadID: 	   int32(threadId),
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

	mappedThreads := make(map[int32][]*threadWithLabels)
	for _, rt := range replyThreads {
		var labels []templates.TopicLabel
		if pub, author, err := cd.ThreadPublicLabels(rt.Idforumthread); err == nil {
			for _, l := range pub {
				labels = append(labels, templates.TopicLabel{Name: l, Type: "public"})
			}
			for _, l := range author {
				labels = append(labels, templates.TopicLabel{Name: l, Type: "author"})
			}
		}
		if priv, err := cd.ThreadPrivateLabels(rt.Idforumthread, rt.Lastposter); err == nil {
			for _, l := range priv {
				labels = append(labels, templates.TopicLabel{Name: l, Type: "private"})
			}
		}

		commentID := int32(0)
		if rt.ReplyToCommentID.Valid {
			commentID = rt.ReplyToCommentID.Int32
		}

		mappedThreads[commentID] = append(mappedThreads[commentID], &threadWithLabels{
			GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow: db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{
				Idforumthread:          rt.Idforumthread,
				Firstpost:              rt.Firstpost,
				Lastposter:             rt.Lastposter,
				ForumtopicIdforumtopic: rt.ForumtopicIdforumtopic,
				Comments:               rt.TotalComments,
				Lastaddition:           rt.Lastaddition,
				Locked:                 rt.Locked,
				DeletedAt:              rt.DeletedAt,
				Lastposterusername:     rt.LastPosterName,
				Firstpostuserid:        sql.NullInt32{},
				Firstpostwritten:       sql.NullTime{},
				Firstposttext:          rt.FirstPostText,
			},
			Labels: labels,
		})
	}

	cd.PageTitle = "Replies - " + topic.Title.String
	data := Data{
		Topic:        topic,
		Thread:       thread,
		ThreadsByComment: mappedThreads,
	}

	if err := ReplyThreadsPageTmpl.Handle(w, r, data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
