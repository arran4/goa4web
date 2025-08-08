package writings

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"
)

// SubmitWritingTask encapsulates creating a new writing.
type SubmitWritingTask struct{ tasks.TaskString }

var submitWritingTask = &SubmitWritingTask{TaskString: TaskSubmitWriting}

var _ tasks.Task = (*SubmitWritingTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*SubmitWritingTask)(nil)
var _ notif.GrantsRequiredProvider = (*SubmitWritingTask)(nil)

func (SubmitWritingTask) Page(w http.ResponseWriter, r *http.Request) { ArticleAddPage(w, r) }

func (SubmitWritingTask) Action(w http.ResponseWriter, r *http.Request) any {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["category"])
	if err != nil {
		return fmt.Errorf("category id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
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
	uid, _ := session.Values["UID"].(int32)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	allowed, err := UserCanCreateWriting(r.Context(), queries, int32(categoryID), uid)
	if err != nil {
		return fmt.Errorf("UserCanCreateWriting fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if !allowed {
		return fmt.Errorf("UserCanCreateWriting deny %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("forbidden")))
	}

	articleID, err := queries.CreateWritingForWriter(r.Context(), db.CreateWritingForWriterParams{
		WriterID:          uid,
		WritingCategoryID: int32(categoryID),
		Title:             sql.NullString{Valid: true, String: title},
		Abstract:          sql.NullString{Valid: true, String: abstract},
		Writing:           sql.NullString{Valid: true, String: body},
		Private:           sql.NullBool{Valid: true, Bool: private},
		LanguageID:        sql.NullInt32{Int32: int32(languageID), Valid: true},
		GrantCategoryID:   sql.NullInt32{Int32: int32(categoryID), Valid: true},
		GranteeID:         sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		return fmt.Errorf("create writing fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if articleID == 0 {
		return fmt.Errorf("create writing deny %w", handlers.ErrRedirectOnSamePageHandler(handlers.ErrForbidden))
	}

	var author string
	if u, err := queries.SystemGetUserByID(r.Context(), uid); err == nil {
		author = u.Username.String
	} else {
		return fmt.Errorf("get user fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["Title"] = title
		evt.Data["Author"] = author
		evt.Data["target"] = notif.Target{Type: "writing", ID: int32(articleID)}
	}

	fullText := strings.Join([]string{abstract, title, body}, " ")
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeWriting, ID: int32(articleID), Text: fullText}
	}

	return handlers.RedirectHandler(fmt.Sprintf("/writings/article/%d", articleID))
}

func (SubmitWritingTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("writingEmail")
}

func (SubmitWritingTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("writing")
	return &s
}

func (SubmitWritingTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}
