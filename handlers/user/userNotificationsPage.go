package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type DismissTask struct{ tasks.TaskString }

var dismissTask = &DismissTask{TaskString: tasks.TaskString(TaskDismiss)}
var _ tasks.Task = (*DismissTask)(nil)

var (
	commentAnchorRegexp = regexp.MustCompile(`#c(\d+)$`)
	threadPathRegexp    = regexp.MustCompile(`(?:/private)?/topic/(\d+)/thread/(\d+)`)
)

func userNotificationsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Notifications"
	handlers.DisableCaching(w)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	if cd.FeedsEnabled {
		cd.RSSFeedURL = cd.GenerateFeedURL("/usr/notifications/rss")
		cd.RSSFeedTitle = "Notifications RSS Feed"
		cd.AtomFeedURL = cd.GenerateFeedURL("/usr/notifications/atom")
		cd.AtomFeedTitle = "Notifications Atom Feed"
	}
	ps := cd.PageSize()
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	showAll := r.URL.Query().Get("all") == "1"

	count, err := cd.InboxNotificationCount(showAll)
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	base := "/usr/notifications"
	pagBase := base
	if showAll {
		pagBase += "?all=1"
	}
	cd.Pagination = &common.OffsetPagination{
		TotalItems: int(count),
		PageSize:   ps,
		Offset:     offset,
		BaseURL:    pagBase,
	}

	pref, _ := cd.UserSettings(cd.UserID)
	var digestHour *int32
	var weeklyDay, weeklyHour *int32
	var monthlyDay, monthlyHour *int32
	var digestMarkRead bool
	var timezone string
	var currentTime string

	if pref != nil {
		if pref.DailyDigestHour.Valid {
			digestHour = &pref.DailyDigestHour.Int32
		}
		if pref.WeeklyDigestDay.Valid {
			weeklyDay = &pref.WeeklyDigestDay.Int32
		}
		if pref.WeeklyDigestHour.Valid {
			weeklyHour = &pref.WeeklyDigestHour.Int32
		}
		if pref.MonthlyDigestDay.Valid {
			monthlyDay = &pref.MonthlyDigestDay.Int32
		}
		if pref.MonthlyDigestHour.Valid {
			monthlyHour = &pref.MonthlyDigestHour.Int32
		}
		digestMarkRead = pref.DailyDigestMarkRead
		if pref.Timezone.Valid {
			timezone = pref.Timezone.String
		}
	}

	dHour := -1
	dEnabled := false
	if digestHour != nil {
		dHour = int(*digestHour)
		dEnabled = true
	}

	wDay := -1
	wHour := 0
	wEnabled := false
	if weeklyDay != nil {
		wDay = int(*weeklyDay)
		wEnabled = true
	}
	if weeklyHour != nil {
		wHour = int(*weeklyHour)
	}

	mDay := -1
	mHour := 0
	mEnabled := false
	if monthlyDay != nil {
		mDay = int(*monthlyDay)
		mEnabled = true
	}
	if monthlyHour != nil {
		mHour = int(*monthlyHour)
	}

	now := time.Now().UTC()
	if timezone != "" {
		if loc, err := time.LoadLocation(timezone); err == nil {
			now = now.In(loc)
			currentTime = now.Format("15:04 MST")
		} else {
			currentTime = now.Format("15:04 MST")
		}
	} else {
		currentTime = now.Format("15:04 MST")
	}

	data := struct {
		Request        *http.Request
		DigestHour     int
		DigestEnabled  bool
		WeeklyDay      int
		WeeklyHour     int
		WeeklyEnabled  bool
		MonthlyDay     int
		MonthlyHour    int
		MonthlyEnabled bool
		DigestMarkRead bool
		Timezone       string
		CurrentTime    string
	}{
		Request:        r,
		DigestHour:     dHour,
		DigestEnabled:  dEnabled,
		WeeklyDay:      wDay,
		WeeklyHour:     wHour,
		WeeklyEnabled:  wEnabled,
		MonthlyDay:     mDay,
		MonthlyHour:    mHour,
		MonthlyEnabled: mEnabled,
		DigestMarkRead: digestMarkRead,
		Timezone:       timezone,
		CurrentTime:    currentTime,
	}
	_ = UserNotificationsPage.Handle(w, r, data)
}

const UserNotificationsPage tasks.Template = "domains/user/notifications.gohtml"

// UserNotificationsListFragment renders the inner HTML of #notifications-list for live updates.
const UserNotificationsListFragment tasks.Template = "domains/user/notificationsListFragment.gohtml"

// UserNotificationsCountFragment renders the #notif-index link with unread count for live updates.
const UserNotificationsCountFragment tasks.Template = "domains/user/notificationsCountFragment.gohtml"

func notificationsListFragmentPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	ps := cd.PageSize()
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	showAll := r.URL.Query().Get("all") == "1"
	count, err := cd.InboxNotificationCount(showAll)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	cd.Pagination = &common.OffsetPagination{
		TotalItems: int(count),
		PageSize:   ps,
		Offset:     offset,
		BaseURL:    "/usr/notifications",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	data := struct {
		Request *http.Request
	}{
		Request: r,
	}
	if err := UserNotificationsListFragment.TemplateExecute(w, r, data); err != nil {
		log.Printf("render notifications list fragment: %v", err)
	}
}

func notificationsCountFragmentPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	cd.NotificationCount = int32(cd.UnreadNotificationCount())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := UserNotificationsCountFragment.TemplateExecute(w, r, struct{}{}); err != nil {
		log.Printf("render notifications count fragment: %v", err)
	}
}

func (DismissTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return nil
	}
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	var ids []int32
	for _, idStr := range r.PostForm["id"] {
		if id, err := strconv.ParseInt(idStr, 10, 32); err == nil && id > 0 {
			ids = append(ids, int32(id))
		}
	}
	for _, raw := range r.PostForm["group"] {
		var snapshot []int32
		if err := json.Unmarshal([]byte(raw), &snapshot); err == nil {
			ids = append(ids, snapshot...)
		}
	}
	if err := cd.ForUser(uid).DismissNotificationSnapshot(r.Context(), ids); err != nil {
		log.Printf("dismiss notifications: %v", err)
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/notifications"}
}

func notificationsRssPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	var uid int32
	vars := mux.Vars(r)
	if username := vars["username"]; username != "" {
		user, err := handlers.VerifyFeedRequest(r, "/usr/notifications/rss")
		if err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		uid = user.Idusers
	} else {
		session := cd.GetSession()
		uid, _ = session.Values["UID"].(int32)
	}
	handlers.DisableCaching(w)
	entries, err := cd.ForUser(uid).VisibleNotificationEntries(db.ListVisibleNotificationEntriesForListerParams{
		GroupUnread: 1, Limit: int32(cd.Config.PageSizeDefault),
	})
	if err != nil {
		log.Printf("notify feed: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	feed := NotificationsFeed(r, common.RenderedNotifications(entries), cd.SiteTitle)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
}

func notificationsAtomPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	var uid int32
	vars := mux.Vars(r)
	if username := vars["username"]; username != "" {
		user, err := handlers.VerifyFeedRequest(r, "/usr/notifications/atom")
		if err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		uid = user.Idusers
	} else {
		session := cd.GetSession()
		uid, _ = session.Values["UID"].(int32)
	}
	handlers.DisableCaching(w)
	entries, err := cd.ForUser(uid).VisibleNotificationEntries(db.ListVisibleNotificationEntriesForListerParams{
		GroupUnread: 1, Limit: int32(cd.Config.PageSizeDefault),
	})
	if err != nil {
		log.Printf("notify feed: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	feed := NotificationsFeed(r, common.RenderedNotifications(entries), cd.SiteTitle)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
}

func fixNotificationLinkAndGetData(cd *common.CoreData, link string) (string, string, string, error) {
	if link == "" {
		return "", "", "", nil
	}

	isBroken := strings.HasSuffix(link, "/reply")
	matches := threadPathRegexp.FindStringSubmatch(link)

	fixedLink := link
	if isBroken {
		fixedLink = strings.TrimSuffix(link, "/reply") + "#bottom"
	}

	if len(matches) < 3 {
		return fixedLink, "", "", nil
	}

	topicID, _ := strconv.Atoi(matches[1])
	threadID, _ := strconv.Atoi(matches[2])

	threadTitle := ""
	sectionTitle := "Forum"

	topic, err := cd.ForumTopicByID(int32(topicID))
	if err == nil && topic != nil {
		if topic.Handler == "private" {
			sectionTitle = "Private Forum"
		}
	}

	thread, err := cd.ForumThreadByID(int32(threadID))
	if err == nil && thread != nil {
		if cmt, err := cd.CommentByID(thread.Firstpost); err == nil && cmt != nil {
			if cmt.Text.Valid {
				threadTitle = a4code.SnipTextWords(cmt.Text.String, 10)
			}
		}
	}

	return fixedLink, threadTitle, sectionTitle, nil
}

func userNotificationOpenPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Notification"
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	handlers.DisableCaching(w)
	if id <= 0 || int64(id) > 2147483647 {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	actor := cd.ForUser(uid)
	n, err := actor.VisibleNotification(int32(id))
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("notification open: %v", err)
		}
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	redirectURL := ""
	if n.Link.Valid {
		redirectURL = n.Link.String
	}

	fixedLink, threadTitle, sectionTitle, _ := fixNotificationLinkAndGetData(cd, redirectURL)
	if fixedLink != "" {
		redirectURL = fixedLink
	}

	replyPreview := ""
	if redirectURL != "" {
		if commentAnchorRegexp.MatchString(redirectURL) {
			matches := commentAnchorRegexp.FindStringSubmatch(redirectURL)
			if len(matches) > 1 {
				if cid, err := strconv.Atoi(matches[1]); err == nil {
					if cmt, err := cd.CommentByID(int32(cid)); err == nil && cmt != nil && cmt.Text.Valid {
						replyPreview = a4code.SnipTextWords(cmt.Text.String, 20)
					}
				}
			}
		}
	}

	if threadTitle != "" {
		cd.PageTitle = fmt.Sprintf("Notification: %s", threadTitle)
	} else if sectionTitle != "" {
		cd.PageTitle = fmt.Sprintf("Notification: %s", sectionTitle)
	}

	data := struct {
		Request      *http.Request
		Notification *db.Notification
		RedirectURL  string
		TaskName     string
		ReplyPreview string
		ThreadTitle  string
		SectionTitle string
	}{
		Request:      r,
		Notification: n,
		RedirectURL:  redirectURL,
		TaskName:     string(TaskDismiss),
		ReplyPreview: replyPreview,
		ThreadTitle:  threadTitle,
		SectionTitle: sectionTitle,
	}
	_ = UserNotificationOpenPage.Handle(w, r, data)
}

const UserNotificationOpenPage tasks.Template = "domains/user/notificationOpen.gohtml"

func UserNotificationEmailActionPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	idStr := r.FormValue("email_id")
	id, _ := strconv.Atoi(idStr)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	val, _ := queries.GetMaxNotificationPriority(r.Context(), uid)
	var maxPr int32
	switch v := val.(type) {
	case int64:
		maxPr = int32(v)
	case int32:
		maxPr = v
	}
	if id != 0 {
		if err := queries.SetNotificationPriorityForLister(r.Context(), db.SetNotificationPriorityForListerParams{ListerID: uid, NotificationPriority: maxPr + 1, ID: int32(id)}); err != nil {
			log.Printf("set notification priority: %v", err)
		}
	}
	http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
}

func notificationsGoPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	handlers.DisableCaching(w)
	if id <= 0 || int64(id) > 2147483647 {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	actor := cd.ForUser(uid)
	n, err := actor.VisibleNotification(int32(id))
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("notification go: %v", err)
		}
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	if !n.ReadAt.Valid {
		if err := actor.DismissNotification(r.Context(), n.ID); err != nil {
			log.Printf("mark notification read: %v", err)
		}
	}
	link := ""
	if n.Link.Valid && n.Link.String != "" {
		link = n.Link.String
	}
	if link != "" {
		if before, ok := strings.CutSuffix(link, "/reply"); ok {
			link = before + "#bottom"
		}
		http.Redirect(w, r, link, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/usr/notifications/open/%d", n.ID), http.StatusSeeOther)
}
