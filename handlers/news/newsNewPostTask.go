package news

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
)

type NewPostTask struct{ tasks.TaskString }

var newPostTask = &NewPostTask{TaskString: TaskNewPost}

// NewPostTask sends notifications when a new post is created and automatically subscribes the author to future replies.
var (
	_ tasks.Task                                    = (*NewPostTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*NewPostTask)(nil)
	_ notif.AdminEmailTemplateProvider              = (*NewPostTask)(nil)
	_ notif.AutoSubscribeProvider                   = (*NewPostTask)(nil)
)

func (NewPostTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsAddEmail")
}

func (NewPostTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsAddEmail")
	return &v
}

func (NewPostTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("newsAddEmail")
}

func (NewPostTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("news_add")
	return &s
}

// AutoSubscribePath links the newly created post so that any future replies notify the author by default.
// Subscribing the poster ensures they are notified when readers engage with their new thread.
// AutoSubscribePath keeps authors in the loop on new post discussions.
// AutoSubscribePath implements notif.AutoSubscribeProvider. Subscriptions use the thread path derived from postcountworker data when possible.
func (NewPostTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskNewPost), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID), nil
	}
	return string(TaskNewPost), evt.Path, nil
}

func (NewPostTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("languageId parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	if cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData); !cd.HasGrant("news", "post", "post", 0) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		handlers.TaskErrorAcknowledgementPage(w, r)
		return nil
	}
	id, err := queries.CreateNewsPostForWriter(r.Context(), db.CreateNewsPostForWriterParams{
		LanguageID: int32(languageId),
		News:       sql.NullString{String: text, Valid: true},
		WriterID:   uid,
		GranteeID:  sql.NullInt32{Int32: uid, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create news post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	// give the author edit rights on the new post
	if _, err := queries.AdminCreateGrant(r.Context(), db.AdminCreateGrantParams{
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

	if u, err := queries.SystemGetUserByID(r.Context(), uid); err == nil {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["Author"] = u.Username.String
				evt.Data["PostURL"] = cd.AbsoluteURL(fmt.Sprintf("/news/news/%d", id))
			}
		}
	}

	return nil
}
