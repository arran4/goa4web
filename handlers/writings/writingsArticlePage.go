package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	common "github.com/arran4/goa4web/handlers/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	searchutil "github.com/arran4/goa4web/internal/utils/searchutil"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
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
		*corecommon.CoreData
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
		IsAdmin             bool
		Categories          []*db.WritingCategory
		CategoryId          int32
		Offset              int32
		CategoryBreadcrumbs []*db.WritingCategory
		ReplyText           string
	}

	cd := r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData)
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           cd,
		CanReply:           cd.UserID != 0,
		CanEdit:            false,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
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
	queries = r.Context().Value(hcommon.KeyQueries).(*db.Queries)

	writing, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), db.GetWritingByIdForUserDescendingByPublishedDateParams{
		ViewerID:      uid,
		Idwriting:     int32(articleId),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_ = templates.GetCompiledTemplates(r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
			return
		default:
			log.Printf("getWritingByIdForUserDescendingByPublishedDate Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if !cd.HasGrant("writing", "article", "view", writing.Idwriting) {
		_ = templates.GetCompiledTemplates(r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
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
	data.CanEdit = (cd.HasRole("administrator") && cd.AdminMode) || (cd.HasRole("content writer") && data.IsAuthor)
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

	common.TemplateHandler(w, r, "articlePage.gohtml", data)
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

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	uid, _ := session.Values["UID"].(int32)

	post, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), db.GetWritingByIdForUserDescendingByPublishedDateParams{
		ViewerID:      uid,
		Idwriting:     int32(aid),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
			_ = templates.GetCompiledTemplates(cd.Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", cd)
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

	if cd, ok := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData); ok {
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

	if err := hcommon.PostUpdate(r.Context(), queries, pthid, ptid); err != nil {
		log.Printf("Error: postUpdate: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	cid := int64(0)
	wordIds, done := searchutil.SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if searchutil.InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}
	//??? if _, done := SearchWordIdsFromText(w, r, text, queries); done {

	hcommon.TaskDoneAutoRefreshPage(w, r)
}
