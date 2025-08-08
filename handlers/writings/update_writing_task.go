package writings

import (
	"database/sql"
	"fmt"
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.CurrentWriting()
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

	queries := cd.Queries()

	if err := queries.UpdateWritingForWriter(r.Context(), db.UpdateWritingForWriterParams{
		Title:      sql.NullString{Valid: true, String: title},
		Abstract:   sql.NullString{Valid: true, String: abstract},
		Content:    sql.NullString{Valid: true, String: body},
		Private:    sql.NullBool{Valid: true, Bool: private},
		LanguageID: sql.NullInt32{Int32: int32(languageID), Valid: true},
		WritingID:  writing.Idwriting,
		WriterID:   cd.UserID,
		GranteeID:  sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	}); err != nil {
		return fmt.Errorf("update writing fail %w", err)
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

func (UpdateWritingTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("writingUpdateEmail")
}

func (UpdateWritingTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("writing_update")
	return &s
}

func (UpdateWritingTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}
