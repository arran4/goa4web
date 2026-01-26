package writings

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"
)

// UpdateWritingTask applies changes to an article.
type UpdateWritingTask struct{ tasks.TaskString }

var updateWritingTask = &UpdateWritingTask{TaskString: TaskUpdateWriting}

var _ tasks.Task = (*UpdateWritingTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*UpdateWritingTask)(nil)
var _ notif.GrantsRequiredProvider = (*UpdateWritingTask)(nil)
var _ tasks.EmailTemplatesRequired = (*UpdateWritingTask)(nil)

func (UpdateWritingTask) Page(w http.ResponseWriter, r *http.Request) { ArticleEditPage(w, r) }

func (UpdateWritingTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.Article()
	if err != nil || writing == nil {
		return fmt.Errorf("current writing fail %w", err)
	}

	languageID, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	private, err := strconv.ParseBool(r.PostFormValue("isitprivate"))
	if err != nil {
		return fmt.Errorf("private parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")
	raw := r.PostForm["author"]
	labels := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, l := range raw {
		if v := strings.TrimSpace(l); v != "" {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				labels = append(labels, v)
			}
		}
	}

	queries := cd.Queries()

	if err := cd.UpdateWriting(writing, title, abstract, body, private, int32(languageID)); err != nil {
		return fmt.Errorf("update writing fail %w", err)
	}

	if err := cd.SetWritingAuthorLabels(writing.Idwriting, labels); err != nil {
		return fmt.Errorf("set author labels fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			author := ""
			if writing.Writerusername.Valid {
				author = writing.Writerusername.String
			}
			evt.Data["Title"] = title
			evt.Data["Author"] = author
			evt.Data["PostURL"] = cd.AbsoluteURL(fmt.Sprintf("/writings/article/%d", writing.Idwriting))
			evt.Data["target"] = notif.Target{Type: "writing", ID: writing.Idwriting}
		}
	}

	if err := queries.SystemDeleteWritingSearchByWritingID(r.Context(), writing.Idwriting); err != nil {
		return fmt.Errorf("writing search delete fail %w", err)
	}

	fullText := strings.Join([]string{abstract, title, body}, " ")
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeWriting, ID: writing.Idwriting, Text: fullText}
		}
	}

	return nil
}

func (UpdateWritingTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateWritingUpdate.EmailTemplates(), true
}

func (UpdateWritingTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateWritingUpdate.NotificationTemplate()
	return &s
}

func (UpdateWritingTask) EmailTemplatesRequired() []tasks.Page {
	return append(EmailTemplateWritingUpdate.RequiredPages(), NotificationTemplateWritingUpdate.RequiredPages()...)
}

func (UpdateWritingTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}
