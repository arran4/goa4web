package news

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	"github.com/arran4/goa4web/core/templates"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	email "github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	searchutil "github.com/arran4/goa4web/internal/searchutil"
	"github.com/arran4/goa4web/runtimeconfig"
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
		SelectedLanguageId: corelanguage.ResolveDefaultLanguageID(r.Context(), queries, runtimeconfig.AppRuntimeConfig.DefaultLanguage),
	}
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), int32(pid))
	if err != nil {
		log.Printf("getNewsPostByIdWithWriterIdAndThreadCommentCount Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	editingId, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	replyType := r.URL.Query().Get("type")

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		UsersIdusers:  uid,
		ForumthreadID: int32(post.ForumthreadID),
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
		UsersIdusers:  uid,
		Idforumthread: int32(post.ForumthreadID),
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

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)

	editCommentIdString := r.URL.Query().Get("editComment")
	editCommentId, _ := strconv.Atoi(editCommentIdString)
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if uid == row.UsersIdusers {
			editUrl = fmt.Sprintf("?editComment=%d", row.Idcomments)
			editSaveUrl = "?"
			// TODO
			//editUrl = fmt.Sprintf("/forum/topic/%d/thread/%d?comment=%d#edit", topicRow.Idforumtopic, threadId, row.Idcomments)
			//editSaveUrl = fmt.Sprintf("/forum/topic/%d/thread/%d/comment/%d", topicRow.Idforumtopic, threadId, row.Idcomments)
			if commentId != 0 && int32(commentId) == row.Idcomments {
				data.IsReplyable = false
			}
		}

		if int32(commentId) == row.Idcomments {
			switch replyType {
			case "full":
				data.ReplyText = hcommon.ProcessCommentFullQuote(row.Posterusername.String, row.Text.String)
			default:
				data.ReplyText = hcommon.ProcessCommentQuote(row.Posterusername.String, row.Text.String)
			}
		}

		data.Comments = append(data.Comments, &CommentPlus{
			GetCommentsByThreadIdForUserRow: row,
			ShowReply:                       data.CoreData.UserID != 0,
			EditUrl:                         editUrl,
			EditSaveUrl:                     editSaveUrl,
			Editing:                         editCommentId != 0 && int32(editCommentId) == row.Idcomments,
			Offset:                          i + offset,
			Languages:                       nil,
			SelectedLanguageId:              0,
		})
	}

	data.Thread = threadRow
	ann, err := queries.GetLatestAnnouncementByNewsID(r.Context(), post.Idsitenews)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("getLatestAnnouncementByNewsID: %v", err)
	}
	data.Post = &Post{
		GetNewsPostByIdWithWriterIdAndThreadCommentCountRow: post,
		ShowReply: data.CoreData.UserID != 0,
		ShowEdit: data.CoreData.HasRole("writer") ||
			data.CoreData.HasRole("moderator") ||
			data.CoreData.HasRole("administrator"),
		Editing:      editingId == int(post.Idsitenews),
		Announcement: ann,
		IsAdmin:      data.CoreData.HasRole("administrator") && data.CoreData.AdminMode,
	}

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomNewsIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "postPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)

	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), int32(pid))
	if err != nil {
		log.Printf("getNewsPostByIdWithWriterIdAndThreadCommentCount Error: %s", err)
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
	uid, _ := session.Values["UID"].(int32)

	base := "http://" + r.Host
	if runtimeconfig.AppRuntimeConfig.HTTPHostname != "" {
		base = strings.TrimRight(runtimeconfig.AppRuntimeConfig.HTTPHostname, "/")
	}
	endUrl := base + fmt.Sprintf("/news/news/%d", pid)

	provider := email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig)
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

	// TODO
	//if rows, err := queries.SomethingNotifyNews(r.Context(), somethingNotifyNewssParams{
	//	Idusers: uid,
	//	Idnewss: int32(bid),
	//}); err != nil {
	//	log.Printf("Error: listUsersSubscribedToThread: %s", err)
	//} else {
	//	for _, row := range rows {
	//		if err := notifyChange(r.Context(), email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig), row.String, endUrl); err != nil {
	//			log.Printf("Error: notifyChange: %s", err)
	//
	//		}
	//	}
	//}

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
	// TODO verify field names
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["post"])

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
	// TODO verify field names
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

	err = queries.CreateNewsPost(r.Context(), db.CreateNewsPostParams{
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

	if u, err := queries.GetUserById(r.Context(), uid); err == nil {
		if evt, ok := r.Context().Value(hcommon.KeyBusEvent).(*eventbus.Event); ok && evt != nil {
			evt.Item = notif.BlogPostInfo{Author: u.Username.String}
		}
	}

	hcommon.TaskDoneAutoRefreshPage(w, r)
}
