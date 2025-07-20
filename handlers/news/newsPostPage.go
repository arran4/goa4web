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
	common "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	"github.com/arran4/goa4web/core/templates"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	postcountworker "github.com/arran4/goa4web/workers/postcountworker"
	searchworker "github.com/arran4/goa4web/workers/searchworker"
)

type NewsPost struct {
	ShowReply bool
	// ShowEdit is true when the current user can modify the post. Users with
	// the writer, moderator or administrator role are permitted to edit.
	ShowEdit bool
}

type ReplyTask struct{ tasks.TaskString }

var replyTask = &ReplyTask{TaskString: TaskReply}

// ReplyTask hooks into notification and auto subscription systems so readers
// following a news post will see replies and admins are emailed about new
// discussions. This promotes active conversations while giving moderators
// oversight.
// Interface checks with reasoning. Administrators and subscribers receive
// notifications when discussions grow, and commenters are auto-subscribed so
// they know when someone replies.
var _ tasks.Task = (*ReplyTask)(nil)

// news readers subscribed to a post should get email when replies land
// ReplyTask keeps commenters in the loop by notifying thread followers and
// subscribing the author to future replies.
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)

// staff get alerted about all replies via admin templates
var _ notif.AdminEmailTemplateProvider = (*ReplyTask)(nil)

// commenters want to follow conversations they've participated in
// auto-subscribe so readers know when someone replies to their comment
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

func (ReplyTask) IndexType() string { return searchworker.TypeComment }

func (ReplyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = ReplyTask{}

func (ReplyTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("replyEmail")
}

func (ReplyTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("reply")
	return &s
}

func (ReplyTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsReplyEmail")
}

func (ReplyTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsReplyEmail")
	return &v
}

// AutoSubscribePath registers this reply so the author automatically follows
// subsequent comments on the news post.
	// When users reply to a news post we automatically subscribe them so
	// they receive updates to the thread they just engaged with.
// AutoSubscribePath allows commenters to automatically watch for further replies.
// AutoSubscribePath implements notif.AutoSubscribeProvider. A subscription to
// the underlying discussion thread is created using event data when available.
func (ReplyTask) AutoSubscribePath(evt eventbus.Event) (string, string) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskReply), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID)
	}
	return string(TaskReply), evt.Path
}

type EditTask struct{ tasks.TaskString }

var editTask = &EditTask{TaskString: TaskEdit}

var _ tasks.Task = (*EditTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*EditTask)(nil)

func (EditTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsEditEmail")
}

func (EditTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsEditEmail")
	return &v
}

type NewPostTask struct{ tasks.TaskString }

var newPostTask = &NewPostTask{TaskString: TaskNewPost}

// NewPostTask implements notification providers so authors automatically
// subscribe to discussion on their posts and administrators are kept in the
// loop. Subscribers are notified as well, encouraging engagement with freshly
// published content.
// New posts alert subscribers and admins and subscribe the poster to reply
// notifications.
var _ tasks.Task = (*NewPostTask)(nil)

// subscribers to the news feed expect updates when new posts appear
var _ notif.SubscribersNotificationTemplateProvider = (*NewPostTask)(nil)

// admins track all postings so they receive dedicated notifications
var _ notif.AdminEmailTemplateProvider = (*NewPostTask)(nil)

// authors should be subscribed to their post automatically for follow-ups
// new posts should auto-subscribe authors for reply alerts
var _ notif.AutoSubscribeProvider = (*NewPostTask)(nil)

func (NewPostTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsAddEmail")
}

func (NewPostTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsAddEmail")
	return &v
}

func (NewPostTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("newsAddEmail")
}

func (NewPostTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("news_add")
	return &s
}

// AutoSubscribePath links the newly created post so that any future replies
// notify the author by default.
	// Subscribing the poster ensures they are notified when readers engage
	// with their new thread.
// AutoSubscribePath keeps authors in the loop on new post discussions.
// AutoSubscribePath implements notif.AutoSubscribeProvider. Subscriptions use
// the thread path derived from postcountworker data when possible.
func (NewPostTask) AutoSubscribePath(evt eventbus.Event) (string, string) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskNewPost), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID)
	}
	return string(TaskNewPost), evt.Path
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
		*common.CoreData
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

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(common.KeyCoreData).(*common.CoreData),
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
			_ = templates.GetCompiledSiteTemplates(r.Context().Value(common.KeyCoreData).(*common.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
			return
		default:
			log.Printf("GetNewsPostByIdWithWriterIdAndThreadCommentCountForUser Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	if !data.CoreData.HasGrant("news", "post", "view", post.Idsitenews) {
		_ = templates.GetCompiledSiteTemplates(r.Context().Value(common.KeyCoreData).(*common.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
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

	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
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

	handlers.TemplateHandler(w, r, "postPage.gohtml", data)
}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) {
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

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
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

	evt := cd.Event()
	evt.Data["news_url"] = endUrl

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

	if cd, ok := r.Context().Value(common.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: pthid, TopicID: ptid}
		}
	}
	if cd, ok := r.Context().Value(common.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: int32(cid), Text: text}
		}
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (EditTask) Action(w http.ResponseWriter, r *http.Request) {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		r.URL.RawQuery = "error=" + url.QueryEscape(err.Error())
		handlers.TaskErrorAcknowledgementPage(w, r)
		return
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["post"])

	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	if !cd.HasGrant("news", "post", "edit", int32(postId)) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		handlers.TaskErrorAcknowledgementPage(w, r)
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

	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (NewPostTask) Action(w http.ResponseWriter, r *http.Request) {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		r.URL.RawQuery = "error=" + url.QueryEscape(err.Error())
		handlers.TaskErrorAcknowledgementPage(w, r)
		return
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	if cd := r.Context().Value(common.KeyCoreData).(*common.CoreData); !cd.HasGrant("news", "post", "post", 0) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		handlers.TaskErrorAcknowledgementPage(w, r)
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
		if cd, ok := r.Context().Value(common.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["blog"] = notif.BlogPostInfo{Author: u.Username.String}
			}
		}
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}
