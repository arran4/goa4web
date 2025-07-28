package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slices"
)

func ArticlePage(w http.ResponseWriter, r *http.Request) {
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
		*common.CoreData
		Writing             *db.GetWritingByIdForUserDescendingByPublishedDateRow
		CanEdit             bool
		IsAuthor            bool
		CanReply            bool
		UserId              int32
		Languages           []*db.Language
		SelectedLanguageId  int
		Thread              *db.GetThreadLastPosterAndPermsRow
		Comments            []*CommentPlus
		IsReplyable         bool
		Categories          []*db.WritingCategory
		CategoryId          int32
		Offset              int32
		CategoryBreadcrumbs []*db.WritingCategory
		ReplyText           string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Writing"
	queries := cd.Queries()
	data := Data{
		CoreData:           cd,
		CanReply:           cd.UserID != 0,
		CanEdit:            false,
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
		IsReplyable:        true,
	}

	vars := mux.Vars(r)
	articleId, _ := strconv.Atoi(vars["article"])

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid
	queries = r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	writing, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), db.GetWritingByIdForUserDescendingByPublishedDateParams{
		ViewerID:      uid,
		Idwriting:     int32(articleId),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err == nil {
		if writing.Title.Valid {
			cd.PageTitle = fmt.Sprintf("Writing: %s", writing.Title.String)
		} else {
			cd.PageTitle = fmt.Sprintf("Writing %d", writing.Idwriting)
		}
	}
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
			if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return
		default:
			log.Printf("getWritingByIdForUserDescendingByPublishedDate Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if !cd.HasGrant("writing", "article", "view", writing.Idwriting) {
		if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}

	if writing.ForumthreadID == 0 && uid != 0 {
		pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
			String: WritingTopicName,
			Valid:  true,
		})
		var ptid int32
		if errors.Is(err, sql.ErrNoRows) {
			ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
				ForumcategoryIdforumcategory: 0,
				Title:                        sql.NullString{String: WritingTopicName, Valid: true},
				Description:                  sql.NullString{String: WritingTopicDescription, Valid: true},
			})
			if err != nil {
				log.Printf("Error: createForumTopic: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			ptid = int32(ptidi)
		} else if err != nil {
			log.Printf("Error: findForumTopicByTitle: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		} else {
			ptid = pt.Idforumtopic
		}

		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		pthid := int32(pthidi)
		if err := queries.AssignWritingThisThreadId(r.Context(), db.AssignWritingThisThreadIdParams{
			ForumthreadID: pthid,
			Idwriting:     writing.Idwriting,
		}); err != nil {
			log.Printf("Error: assign_article_to_thread: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		writing.ForumthreadID = pthid
	}

	data.Writing = writing
	data.IsAuthor = writing.UsersIdusers == uid
	data.CanEdit = (cd.HasAdminRole() && cd.AdminMode) || (cd.HasContentWriterRole() && data.IsAuthor)
	data.CategoryId = writing.WritingCategoryID

	languageRows, err := cd.Languages()
	if err != nil {
		log.Printf("FetchLanguages Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: uid,
		ThreadID: writing.ForumthreadID,
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
		ThreadID:      writing.ForumthreadID,
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

	categoryRows, err := data.CoreData.WritingCategories()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllWritingCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	categoryMap := map[int32]*db.WritingCategory{}
	for _, cat := range categoryRows {
		categoryMap[cat.Idwritingcategory] = cat
		if cat.WritingCategoryID == data.CategoryId {
			data.Categories = append(data.Categories, cat)
		}
	}
	for cid := data.CategoryId; len(data.CategoryBreadcrumbs) < len(categoryRows); {
		cat, ok := categoryMap[cid]
		if ok {
			data.CategoryBreadcrumbs = append(data.CategoryBreadcrumbs, cat)
			cid = cat.WritingCategoryID
		} else {
			break
		}
	}
	slices.Reverse(data.CategoryBreadcrumbs)

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)
	editCommentIdString := r.URL.Query().Get("editComment")
	editCommentId, _ := strconv.Atoi(editCommentIdString)
	replyType := r.URL.Query().Get("type")
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if data.CoreData.CanEditAny() || row.IsOwner {
			editUrl = fmt.Sprintf("/writings/article/%d?comment=%d#edit", writing.Idwriting, row.Idcomments)
			editSaveUrl = fmt.Sprintf("/writings/article/%d/comment/%d", writing.Idwriting, row.Idcomments)
			if editCommentId != 0 && int32(editCommentId) == row.Idcomments {
				data.IsReplyable = false
			}
		}

		if int32(commentId) == row.Idcomments {
			switch replyType {
			case "full":
				data.ReplyText = a4code.FullQuoteOf(row.Posterusername.String, row.Text.String)
			default:
				data.ReplyText = a4code.QuoteOfText(row.Posterusername.String, row.Text.String)
			}
		}

		data.Comments = append(data.Comments, &CommentPlus{
			GetCommentsByThreadIdForUserRow: row,
			ShowReply:                       data.CoreData.UserID != 0,
			EditUrl:                         editUrl,
			EditSaveUrl:                     editSaveUrl,
			Editing:                         editCommentId != 0 && (data.CoreData.CanEditAny() || row.IsOwner) && int32(editCommentId) == row.Idcomments,
			Offset:                          i + offset,
			Languages:                       languageRows,
			SelectedLanguageId:              row.LanguageIdlanguage,
		})
	}

	data.Thread = threadRow

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

	post, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), db.GetWritingByIdForUserDescendingByPublishedDateParams{
		ViewerID:      uid,
		Idwriting:     int32(aid),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
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
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: WritingTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
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
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		pthid = int32(pthidi)
		if err := queries.AssignWritingThisThreadId(r.Context(), db.AssignWritingThisThreadIdParams{
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

	if _, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
		ForumthreadID:      pthid,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
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
