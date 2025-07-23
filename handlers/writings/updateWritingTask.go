package writings

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
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

func (UpdateWritingTask) Page(w http.ResponseWriter, r *http.Request) { ArticleEditPage(w, r) }

func (UpdateWritingTask) Action(w http.ResponseWriter, r *http.Request) any {
	writing := r.Context().Value(consts.KeyWriting).(*db.GetWritingByIdForUserDescendingByPublishedDateRow)

	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	private, err := strconv.ParseBool(r.PostFormValue("isitprivate"))
	if err != nil {
		return fmt.Errorf("private parse %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	if err := queries.UpdateWriting(r.Context(), db.UpdateWritingParams{
		Title:              sql.NullString{Valid: true, String: title},
		Abstract:           sql.NullString{Valid: true, String: abstract},
		Writing:            sql.NullString{Valid: true, String: body},
		Private:            sql.NullBool{Valid: true, Bool: private},
		LanguageIdlanguage: int32(languageId),
		Idwriting:          writing.Idwriting,
	}); err != nil {
		log.Printf("updateWriting Error: %s", err)
		return fmt.Errorf("update writing %w", err)
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

	if err := queries.WritingSearchDelete(r.Context(), writing.Idwriting); err != nil {
		log.Printf("writingSearchDelete Error: %s", err)
		return fmt.Errorf("delete search index %w", err)
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

func (UpdateWritingTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("writingUpdateEmail")
}

func (UpdateWritingTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("writing_update")
	return &s
}

// GrantsRequired implements notif.GrantsRequiredProvider for article updates.
func (UpdateWritingTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}
