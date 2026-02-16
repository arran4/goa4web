package faq

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/faq_templates"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// CreateTemplateTask creates a FAQ entry from a stored template.
type CreateTemplateTask struct{ tasks.TaskString }

var createTemplateTask = &CreateTemplateTask{TaskString: TaskCreateFromTemplate}
var _ tasks.Task = (*CreateTemplateTask)(nil)

func (CreateTemplateTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskCreateFromTemplate)(r, m)
}

func (CreateTemplateTask) Action(w http.ResponseWriter, r *http.Request) any {
	templateName := r.PostFormValue("template")
	if templateName == "" {
		return fmt.Errorf("missing template name %w", handlers.ErrRedirectOnSamePageHandler(errors.New("template name required")))
	}

	categoryID, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		return fmt.Errorf("category parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	languageID, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	authorID, err := strconv.Atoi(r.PostFormValue("author"))
	if err != nil {
		return fmt.Errorf("author parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasGrant("faq", "question", "post", 0) {
		r.URL.RawQuery = "error=" + url.QueryEscape("Forbidden")
		handlers.TaskErrorAcknowledgementPage(w, r)
		return nil
	}

	if authorID == 0 && cd.UserID != 0 {
		authorID = int(cd.UserID)
	}
	if authorID == 0 {
		return fmt.Errorf("author missing %w", handlers.ErrRedirectOnSamePageHandler(errors.New("author required")))
	}

	content, err := faq_templates.Get(templateName)
	if err != nil {
		return fmt.Errorf("template load fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	version, description, question, answer, err := faq_templates.ParseTemplateContent(content)
	if err != nil {
		return fmt.Errorf("template parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	params := db.AdminCreateFAQParams{
		Question:    sql.NullString{String: question, Valid: question != ""},
		Answer:      sql.NullString{String: answer, Valid: answer != ""},
		CategoryID:  sql.NullInt32{Int32: int32(categoryID), Valid: categoryID != 0},
		AuthorID:    int32(authorID),
		LanguageID:  sql.NullInt32{Int32: int32(languageID), Valid: languageID != 0},
		Priority:    0,
		Description: sql.NullString{String: description, Valid: description != ""},
		Version:     sql.NullString{String: version, Valid: version != ""},
	}

	res, err := cd.Queries().AdminCreateFAQ(r.Context(), params)
	if err != nil {
		return fmt.Errorf("create faq from template fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("create faq from template fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	vals := url.Values{}
	vals.Set("template", templateName)
	vals.Set("category", strconv.Itoa(categoryID))
	vals.Set("language", strconv.Itoa(languageID))
	vals.Set("author", strconv.Itoa(authorID))
	vals.Set("created", strconv.FormatInt(id, 10))
	return handlers.RedirectHandler("/admin/faq/templates?" + vals.Encode())
}
