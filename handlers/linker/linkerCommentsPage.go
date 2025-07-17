package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	searchworker "github.com/arran4/goa4web/internal/searchworker"

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

func CommentsPage(w http.ResponseWriter, r *http.Request) {
	type CommentPlus struct {
		*db.GetCommentsByThreadIdForUserRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Languages          []*db.Language
		SelectedLanguageId int32
		EditSaveUrl        string
	}
	type Data struct {
		*corecommon.CoreData
		Link               *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
		CanReply           bool
		Languages          []*db.Language
		Comments           []*CommentPlus
		SelectedLanguageId int
		IsReplyable        bool
		Offset             int
		Text               string
		CanEdit            bool
		UserId             int32
		Thread             *db.GetThreadLastPosterAndPermsRow
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	cd := r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData)
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           cd,
		CanReply:           cd.UserID != 0,
		CanEdit:            false,
		Offset:             offset,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
	}
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	queries = r.Context().Value(handlers.KeyQueries).(*db.Queries)

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(r.Context(), int32(linkId))
	if err != nil {
		log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Link = link
	data.CanEdit = cd.HasRole("administrator") || uid == link.UsersIdusers

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: uid,
		ThreadID: link.ForumthreadID,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getBlogEntryForUserById_comments Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      link.ForumthreadID,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Error: getThreadByIdForUserByIdWithLastPosterUserNameAndPermissions: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if data.CoreData.CanEditAny() || row.IsOwner {
			editUrl = fmt.Sprintf("/linker/comments/%d?comment=%d#edit", link.Idlinker, row.Idcomments)
			editSaveUrl = fmt.Sprintf("/linker/comments/%d/comment/%d", link.Idlinker, row.Idcomments)
			if commentId != 0 && int32(commentId) == row.Idcomments {
				data.IsReplyable = false
			}
		}

		data.Comments = append(data.Comments, &CommentPlus{
			GetCommentsByThreadIdForUserRow: row,
			ShowReply:                       true,
			EditUrl:                         editUrl,
			EditSaveUrl:                     editSaveUrl,
			Editing:                         commentId != 0 && int32(commentId) == row.Idcomments,
			Offset:                          i + offset,
			Languages:                       languageRows,
			SelectedLanguageId:              row.LanguageIdlanguage,
		})
	}

	data.Thread = threadRow

	handlers.TemplateHandler(w, r, "commentsPage.gohtml", data)
}

type replyTask struct{ tasks.BasicTaskEvent }

var ReplyTaskEvent = replyTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskReply,
		Match:     tasks.HasTask(TaskReply),
	},
}

func (replyTask) Action(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	linkId, err := strconv.Atoi(vars["link"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if linkId == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(r.Context(), int32(linkId))
	if err != nil {
		log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pthid int32 = link.ForumthreadID
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: LinkerTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: LinkerTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: LinkerTopicDescription,
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
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		pthid = int32(pthidi)
		if err := queries.AssignLinkerThisThreadId(r.Context(), db.AssignLinkerThisThreadIdParams{
			ForumthreadID: pthid,
			Idlinker:      int32(linkId),
		}); err != nil {
			log.Printf("Error: assignThreadIdToBlogEntry: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/linker/comments/%d", linkId)
	cid, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
		ForumthreadID:      pthid,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		log.Printf("Error: createComment: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := PostUpdate(r.Context(), queries, pthid, ptid); err != nil {
		log.Printf("Error: postUpdate: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if cd, ok := r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: cid, Text: text}
		}
	}

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
