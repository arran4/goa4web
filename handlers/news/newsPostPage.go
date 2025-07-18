package news

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	email "github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/utils/emailutil"
	searchutil "github.com/arran4/goa4web/internal/utils/searchutil"
)

type NewsPost struct {
	ShowReply bool
	// ShowEdit is true when the current user can modify the post. Users with
	// the writer, moderator or administrator role are permitted to edit.
	ShowEdit bool
}

func NewsPostPage(w http.ResponseWriter, r *http.Request) {
	type CommentPlus struct {
		*db.GetCommentsByThreadIdForUserRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Languages          []*db.Language
		SelectedLanguageId int
		EditSaveUrl        string
	}
	type Post struct {
		*db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
		ShowReply    bool
		ShowEdit     bool
		Editing      bool
		Announcement *db.SiteAnnouncement
		IsAdmin      bool
	}
	type Data struct {
		*hcommon.CoreData
		Post               *Post
		Languages          []*db.Language
		SelectedLanguageId int32
		Topic              *db.Forumtopic
		Comments           []*CommentPlus
		Offset             int
		IsReplying         bool
		IsReplyable        bool
		Thread             *db.GetThreadLastPosterAndPermsRow
		ReplyText          string
	}

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
		IsReplying:         r.URL.Query().Has("comment"),
		IsReplyable:        true,
		SelectedLanguageId: corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage),
	}
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: uid,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_ = templates.GetCompiledTemplates(r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
			return
		default:
			log.Printf("GetNewsPostByIdWithWriterIdAndThreadCommentCountForUser Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	if !data.CoreData.HasGrant("news", "post", "view", post.Idsitenews) {
		_ = templates.GetCompiledTemplates(r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
		return
	}

	editingId, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	replyType := r.URL.Query().Get("type")

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: uid,
		ThreadID: int32(post.ForumthreadID),
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
		ThreadID:      int32(post.ForumthreadID),
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

	cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)

	editCommentIdString := r.URL.Query().Get("editComment")
	editCommentId, _ := strconv.Atoi(editCommentIdString)
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if data.CoreData.CanEditAny() || row.IsOwner {
			editUrl = fmt.Sprintf("?editComment=%d#edit", row.Idcomments)
			editSaveUrl = fmt.Sprintf("/news/news/%d/comment/%d", pid, row.Idcomments)
			if commentId != 0 && int32(commentId) == row.Idcomments {
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
			Editing:                         editCommentId != 0 && int32(editCommentId) == row.Idcomments,
			Offset:                          i + offset,
			Languages:                       languageRows,
			SelectedLanguageId:              int(row.LanguageIdlanguage),
		})
	}

	data.Thread = threadRow
	ann, err := data.CoreData.NewsAnnouncement(post.Idsitenews)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("announcementForNews: %v", err)
	}
	data.Post = &Post{
		GetNewsPostByIdWithWriterIdAndThreadCommentCountRow: post,
		ShowReply:    data.CoreData.UserID != 0,
		ShowEdit:     canEditNewsPost(data.CoreData, post.Idsitenews),
		Editing:      editingId == int(post.Idsitenews),
		Announcement: ann,
		IsAdmin:      data.CoreData.HasRole("administrator") && data.CoreData.AdminMode,
	}

	common.TemplateHandler(w, r, "postPage.gohtml", data)
}

func NewsPostReplyActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	pid, err := strconv.Atoi(vars["post"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if pid == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	if !cd.HasGrant("news", "post", "reply", int32(pid)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: uid,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		log.Printf("GetNewsPostByIdWithWriterIdAndThreadCommentCountForUser Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pthid = post.ForumthreadID
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: NewsTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: NewsTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: NewsTopicDescription,
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
		if err := queries.AssignNewsThisThreadId(r.Context(), db.AssignNewsThisThreadIdParams{
			ForumthreadID: pthid,
			Idsitenews:    int32(pid),
		}); err != nil {
			log.Printf("Error: assign_news_to_thread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	base := "http://" + r.Host
	if config.AppRuntimeConfig.HTTPHostname != "" {
		base = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/")
	}
	endUrl := base + fmt.Sprintf("/news/news/%d", pid)

	provider := email.ProviderFromConfig(config.AppRuntimeConfig)
	var author string
	if u, err := queries.GetUserById(r.Context(), uid); err == nil {
		author = u.Username.String
	}
	action := "comment"
	if author != "" {
		action = fmt.Sprintf("comment by %s", author)
	}

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), db.ListUsersSubscribedToThreadParams{
		ForumthreadID: pthid,
		Idusers:       uid,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, row.Idusers, row.Email, endUrl, action, nil); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	emailutil.NotifyNewsSubscribers(r.Context(), queries, int32(pid), uid, endUrl)

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

	if err := PostUpdateLocal(r.Context(), queries, pthid, ptid); err != nil {
		log.Printf("Error: postUpdate: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	wordIds, done := searchutil.SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}
	if searchutil.InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}

	hcommon.TaskDoneAutoRefreshPage(w, r)
}

func NewsPostEditActionPage(w http.ResponseWriter, r *http.Request) {
	if err := hcommon.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		r.URL.RawQuery = "error=" + url.QueryEscape(err.Error())
		hcommon.TaskErrorAcknowledgementPage(w, r)
		return
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["post"])

	cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	if !cd.HasGrant("news", "post", "edit", int32(postId)) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		hcommon.TaskErrorAcknowledgementPage(w, r)
		return
	}
	err = queries.UpdateNewsPost(r.Context(), db.UpdateNewsPostParams{
		Idsitenews:         int32(postId),
		LanguageIdlanguage: int32(languageId),
		News: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	hcommon.TaskDoneAutoRefreshPage(w, r)
}

func NewsPostNewActionPage(w http.ResponseWriter, r *http.Request) {
	if err := hcommon.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		r.URL.RawQuery = "error=" + url.QueryEscape(err.Error())
		hcommon.TaskErrorAcknowledgementPage(w, r)
		return
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	if cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData); !cd.HasGrant("news", "post", "post", 0) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		hcommon.TaskErrorAcknowledgementPage(w, r)
		return
	}
	id, err := queries.CreateNewsPost(r.Context(), db.CreateNewsPostParams{
		LanguageIdlanguage: int32(languageId),
		News: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers: uid,
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	// give the author edit rights on the new post
	if _, err := queries.CreateGrant(r.Context(), db.CreateGrantParams{
		UserID:   sql.NullInt32{Int32: uid, Valid: true},
		RoleID:   sql.NullInt32{},
		Section:  "news",
		Item:     sql.NullString{String: "post", Valid: true},
		RuleType: "allow",
		ItemID:   sql.NullInt32{Int32: int32(id), Valid: true},
		ItemRule: sql.NullString{},
		Action:   "edit",
		Extra:    sql.NullString{},
	}); err != nil {
		log.Printf("create grant: %v", err)
	}

	if u, err := queries.GetUserById(r.Context(), uid); err == nil {
		if cd, ok := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["blog"] = notifications.BlogPostInfo{Author: u.Username.String}
			}
		}
	}

	hcommon.TaskDoneAutoRefreshPage(w, r)
}
