package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/gorilla/mux"
)

func ArticlePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Request            *http.Request
		Writing            *db.GetWritingForListerByIDRow
		Comments           []*db.GetCommentsByThreadIdForUserRow
		CanEdit            bool
		IsAuthor           bool
		CanReply           bool
		Languages          []*db.Language
		SelectedLanguageId int
		ReplyText          string
		EditCommentID      int32
		Offset             int
		IsReplyable        bool
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Writing"

	vars := mux.Vars(r)
	articleId, _ := strconv.Atoi(vars["article"])
	cd.SetCurrentWriting(int32(articleId))
	writing, err := cd.CurrentWriting()
	if err != nil {
		log.Printf("get writing: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if writing == nil || !cd.HasGrant("writing", "article", "view", writing.Idwriting) {
		if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}
	if writing.Title.Valid {
		cd.PageTitle = fmt.Sprintf("Writing: %s", writing.Title.String)
	} else {
		cd.PageTitle = fmt.Sprintf("Writing %d", writing.Idwriting)
	}

	languages, err := cd.Languages()
	if err != nil {
		log.Printf("languages: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	editCommentId, _ := strconv.Atoi(r.URL.Query().Get("editComment"))
	replyType := r.URL.Query().Get("type")

	comments, err := cd.ThreadComments(writing.ForumthreadID)
	if err != nil {
		log.Printf("thread comments: %v", err)
	}
	data := Data{
		Request:            r,
		Writing:            writing,
		Comments:           comments,
		CanReply:           cd.UserID != 0,
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
		Languages:          languages,
		EditCommentID:      int32(editCommentId),
		Offset:             offset,
		IsReplyable:        editCommentId == 0,
	}

	data.IsAuthor = writing.UsersIdusers == cd.UserID
	data.CanEdit = (cd.HasAdminRole() && cd.AdminMode) || (cd.HasContentWriterRole() && data.IsAuthor)

	if c, err := cd.CurrentComment(r); err == nil && c != nil {
		switch replyType {
		case "full":
			data.ReplyText = a4code.FullQuoteOf(c.Username.String, c.Text.String)
		default:
			data.ReplyText = a4code.QuoteOfText(c.Username.String, c.Text.String)
		}
	}

	handlers.TemplateHandler(w, r, "articlePage.gohtml", data)
}

func ArticleReplyActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	aid, err := strconv.Atoi(vars["article"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if aid == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	uid, _ := session.Values["UID"].(int32)

	post, err := queries.GetWritingForListerByID(r.Context(), db.GetWritingForListerByIDParams{
		ListerID:      uid,
		Idwriting:     int32(aid),
		ListerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
			if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return
		default:
			log.Printf("getArticlePost Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	var pthid int32 = post.ForumthreadID
	pt, err := queries.SystemGetForumTopicByTitle(r.Context(), sql.NullString{
		String: WritingTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.SystemCreateForumTopic(r.Context(), db.SystemCreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: WritingTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: WritingTopicDescription,
				Valid:  true,
			},
		})
		if err != nil {
			log.Printf("Error: createForumTopic: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("Error: findForumTopicByTitle: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	} else {
		ptid = pt.Idforumtopic
	}
	if pthid == 0 {
		pthidi, err := queries.SystemCreateThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		pthid = int32(pthidi)
		if err := queries.SystemAssignWritingThreadID(r.Context(), db.SystemAssignWritingThreadIDParams{
			ForumthreadID: pthid,
			Idwriting:     int32(aid),
		}); err != nil {
			log.Printf("Error: assign_article_to_thread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["target"] = notifications.Target{Type: "writing", ID: int32(aid)}
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	if _, err := queries.CreateCommentForCommenter(r.Context(), db.CreateCommentForCommenterParams{
		LanguageID:         int32(languageId),
		CommenterID:        uid,
		ForumthreadID:      pthid,
		Text:               sql.NullString{String: text, Valid: true},
		GrantForumthreadID: sql.NullInt32{Int32: pthid, Valid: true},
		GranteeID:          sql.NullInt32{Int32: uid, Valid: true},
	}); err != nil {
		log.Printf("Error: createComment: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: pthid, TopicID: ptid}
		}
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: 0, Text: text}
		}
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}
