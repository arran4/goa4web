package user

import (
	"database/sql"
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

	"github.com/arran4/goa4web/core"
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
	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	ps := cd.PageSize()
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	showAll := r.URL.Query().Get("all") == "1"

	var count int64
	var err error
	if showAll {
		count, err = cd.Queries().GetNotificationCountForLister(r.Context(), cd.UserID)
	} else {
		count, err = cd.Queries().GetUnreadNotificationCountForLister(r.Context(), cd.UserID)
	}
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	numPages := int((count + int64(ps) - 1) / int64(ps))
	currentPage := offset/ps + 1
	base := "/usr/notifications"
	allParam := ""
	if showAll {
		allParam = "&all=1"
	}
	for i := 1; i <= numPages; i++ {
		cd.PageLinks = append(cd.PageLinks, common.PageLink{
			Num:    i,
			Link:   fmt.Sprintf("%s?offset=%d%s", base, (i-1)*ps, allParam),
			Active: i == currentPage,
		})
	}
	if offset+ps < int(count) {
		cd.NextLink = fmt.Sprintf("%s?offset=%d%s", base, offset+ps, allParam)
	}
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("%s?offset=%d%s", base, offset-ps, allParam)
		cd.StartLink = fmt.Sprintf("%s?offset=0%s", base, allParam)
	}

	pref, _ := cd.UserSettings(cd.UserID)
	var digestHour *int32
	var digestMarkRead bool
	var timezone string
	var currentTime string

	if pref != nil {
		if pref.DailyDigestHour.Valid {
			digestHour = &pref.DailyDigestHour.Int32
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
		DigestMarkRead bool
		Timezone       string
		CurrentTime    string
	}{
		Request:        r,
		DigestHour:     dHour,
		DigestEnabled:  dEnabled,
		DigestMarkRead: digestMarkRead,
		Timezone:       timezone,
		CurrentTime:    currentTime,
	}
	UserNotificationsPage.Handle(w, r, data)
}

const UserNotificationsPage tasks.Template = "user/notifications.gohtml"

func (DismissTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return nil
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	ids := r.Form["id"]
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	for _, idStr := range ids {
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			continue
		}
		n, err := queries.GetNotificationForLister(r.Context(), db.GetNotificationForListerParams{ID: int32(id), ListerID: uid})
		if err == nil && !n.ReadAt.Valid {
			if err := queries.SetNotificationReadForLister(r.Context(), db.SetNotificationReadForListerParams{ID: n.ID, ListerID: uid}); err != nil {
				log.Printf("mark notification read: %v", err)
			}
		}
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
		session, ok := core.GetSessionOrFail(w, r)
		if !ok {
			return
		}
		uid, _ = session.Values["UID"].(int32)
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	limit := int32(cd.Config.PageSizeDefault)
	notifs, err := queries.ListUnreadNotificationsForLister(r.Context(), db.ListUnreadNotificationsForListerParams{ListerID: uid, Limit: limit, Offset: 0})
	if err != nil {
		log.Printf("notify feed: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	feed := NotificationsFeed(r, notifs, cd.SiteTitle)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
		session, ok := core.GetSessionOrFail(w, r)
		if !ok {
			return
		}
		uid, _ = session.Values["UID"].(int32)
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	limit := int32(cd.Config.PageSizeDefault)
	notifs, err := queries.ListUnreadNotificationsForLister(r.Context(), db.ListUnreadNotificationsForListerParams{ListerID: uid, Limit: limit, Offset: 0})
	if err != nil {
		log.Printf("notify feed: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	feed := NotificationsFeed(r, notifs, cd.SiteTitle)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	n, err := queries.GetNotificationForLister(r.Context(), db.GetNotificationForListerParams{ID: int32(id), ListerID: uid})
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
	UserNotificationOpenPage.Handle(w, r, data)
}

const UserNotificationOpenPage tasks.Template = "user/notificationOpen.gohtml"

func userNotificationEmailActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
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
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	n, err := queries.GetNotificationForLister(r.Context(), db.GetNotificationForListerParams{ID: int32(id), ListerID: uid})
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("notification go: %v", err)
		}
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	if !n.ReadAt.Valid {
		if err := queries.SetNotificationReadForLister(r.Context(), db.SetNotificationReadForListerParams{ID: n.ID, ListerID: uid}); err != nil {
			log.Printf("mark notification read: %v", err)
		}
	}
	link := ""
	if n.Link.Valid && n.Link.String != "" {
		link = n.Link.String
	}
	if link != "" {
		if strings.HasSuffix(link, "/reply") {
			link = strings.TrimSuffix(link, "/reply") + "#bottom"
		}
		http.Redirect(w, r, link, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/usr/notifications/open/%d", n.ID), http.StatusSeeOther)
}
