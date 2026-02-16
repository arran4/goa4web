package faq

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/faq_templates"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminTemplatesPageData struct {
	Templates          []string
	SelectedTemplate   string
	Description        string
	Question           string
	Answer             string
	Categories         []*db.FaqCategory
	Languages          []*db.Language
	Authors            []*db.AdminListAllUsersRow
	SelectedCategoryID int32
	SelectedLanguageID int32
	SelectedAuthorID   int32
	CreatedFAQID       int64
	CreatedFAQURL      string
}

// AdminTemplatesPageTmpl renders the admin FAQ templates page.
const AdminTemplatesPageTmpl tasks.Template = "faq/adminTemplatesPage.gohtml"

// AdminTemplatesPage renders the templates administration view.
func AdminTemplatesPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "FAQ Templates"

	names, err := faq_templates.List()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	sort.Strings(names)

	selected := r.URL.Query().Get("template")

	var question string
	var answer string
	var description string
	if selected != "" {
		content, err := faq_templates.Get(selected)
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		description, question, answer, err = faq_templates.ParseTemplateContent(content)
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
	}

	categories, err := cd.FAQCategories()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	languages, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	authors, err := cd.AdminListUsers()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	selectedCategoryID := int32(0)
	if categoryParam := r.URL.Query().Get("category"); categoryParam != "" {
		if value, err := strconv.Atoi(categoryParam); err == nil {
			selectedCategoryID = int32(value)
		}
	}
	selectedLanguageID := cd.PreferredLanguageID(cd.Config.DefaultLanguage)
	if languageParam := r.URL.Query().Get("language"); languageParam != "" {
		if value, err := strconv.Atoi(languageParam); err == nil {
			selectedLanguageID = int32(value)
		}
	}
	selectedAuthorID := cd.UserID
	if authorParam := r.URL.Query().Get("author"); authorParam != "" {
		if value, err := strconv.Atoi(authorParam); err == nil {
			selectedAuthorID = int32(value)
		}
	}

	var createdID int64
	createdParam := r.URL.Query().Get("created")
	if createdParam != "" {
		if value, err := strconv.ParseInt(createdParam, 10, 64); err == nil {
			createdID = value
		}
	}
	createdURL := ""
	if createdID != 0 {
		createdURL = fmt.Sprintf("/admin/faq/question/%d/edit", createdID)
	}

	data := AdminTemplatesPageData{
		Templates:          names,
		SelectedTemplate:   selected,
		Description:        description,
		Question:           question,
		Answer:             answer,
		Categories:         categories,
		Languages:          languages,
		Authors:            authors,
		SelectedCategoryID: selectedCategoryID,
		SelectedLanguageID: selectedLanguageID,
		SelectedAuthorID:   selectedAuthorID,
		CreatedFAQID:       createdID,
		CreatedFAQURL:      createdURL,
	}

	if err := AdminTemplatesPageTmpl.Handle(w, r, data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
