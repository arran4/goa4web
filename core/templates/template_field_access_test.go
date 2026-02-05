package templates_test

import (
	"database/sql"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

func TestTemplateFieldAccess(t *testing.T) {
	// Define test cases mapping template names to sample data
	cases := []struct {
		Name string
		Data any
	}{
		{
			Name: "faq/faqAdminCategoryEditPage.gohtml",
			Data: struct {
				Category   *db.AdminGetFAQCategoryWithQuestionCountByIDRow
				Categories []*db.FaqCategory
			}{
				Category: &db.AdminGetFAQCategoryWithQuestionCountByIDRow{
					ID:            1,
					Name:          sql.NullString{String: "Test Category", Valid: true},
					Questioncount: 5,
				},
				Categories: []*db.FaqCategory{},
			},
		},
		{
			Name: "faq/faqAdminCategoryPage.gohtml",
			Data: struct {
				Category  *db.AdminGetFAQCategoryWithQuestionCountByIDRow
				Latest    []*db.Faq
				Templates []string
				Grants    []*db.SearchGrantsRow
				Roles     []*db.Role
			}{
				Category: &db.AdminGetFAQCategoryWithQuestionCountByIDRow{
					ID:            1,
					Name:          sql.NullString{String: "Test Category", Valid: true},
					Questioncount: 5,
				},
				Latest: []*db.Faq{
					{
						ID:       101,
						Question: sql.NullString{String: "Question 1", Valid: true},
					},
				},
				Templates: []string{"test"},
				Grants:    []*db.SearchGrantsRow{},
				Roles:     []*db.Role{},
			},
		},
		{
			Name: "faq/faqAdminCategoryQuestionsPage.gohtml",
			Data: struct {
				Category  *db.AdminGetFAQCategoryWithQuestionCountByIDRow
				Questions []*db.Faq
			}{
				Category: &db.AdminGetFAQCategoryWithQuestionCountByIDRow{
					ID:            1,
					Name:          sql.NullString{String: "Test Category", Valid: true},
					Questioncount: 5,
				},
				Questions: []*db.Faq{
					{
						ID:       102,
						Question: sql.NullString{String: "Question 2", Valid: true},
					},
				},
			},
		},
	}

	// Setup necessary context
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := common.GetTemplateFuncs(cd, r)
	if _, ok := funcs["assetHash"]; !ok {
		funcs["assetHash"] = func(s string) string { return s }
	}

	// Load templates
	tpl := templates.GetCompiledSiteTemplates(funcs, templates.WithSilence(true))

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tpl.ExecuteTemplate(io.Discard, tc.Name, tc.Data)
			if err != nil {
				t.Errorf("failed to execute template %s: %v", tc.Name, err)
			}
		})
	}
}
