package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
)

// ReplyTask posts a comment reply.
type ReplyTask struct{ tasks.TaskString }

var replyTask = &ReplyTask{TaskString: TaskReply}

var _ tasks.Task = (*ReplyTask)(nil)
var _ notif.GrantsRequiredProvider = (*ReplyTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)
var _ searchworker.IndexedTask = ReplyTask{}

func (ReplyTask) IndexType() string { return searchworker.TypeComment }

func (ReplyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

func (ReplyTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	return notif.NewEmailTemplates("replyEmail")
}

func (ReplyTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	s := notif.NotificationTemplateFilenameGenerator("reply")
	return &s
}

func (ReplyTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}

func (ReplyTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskReply), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID), nil
	}
	return string(TaskReply), evt.Path, nil
}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return handlers.SessionFetchFail{}
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	writing, err := cd.Article()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{}); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return nil
		}
		return fmt.Errorf("get writing fail %w", err)
	}
	if writing == nil {
		return fmt.Errorf("get writing fail %w", handlers.ErrRedirectOnSamePageHandler(sql.ErrNoRows))
	}

	if !cd.HasGrant("writing", "article", "reply", writing.Idwriting) {
		return handlers.ErrRedirectOnSamePageHandler(handlers.ErrForbidden)
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["target"] = notif.Target{Type: "writing", ID: writing.Idwriting}
		}
	}

	text := r.PostFormValue("replytext")
	languageID, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cid, threadID, topicID, err := cd.CreateWritingReply(writing, int32(languageID), text)
	if err != nil {
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cid == 0 {
		err := handlers.ErrForbidden
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: int32(cid), ThreadID: threadID, TopicID: topicID}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: int32(cid), Text: text}
		}
	}

	return nil
}
