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
				Category *db.AdminGetFAQCategoryWithQuestionCountByIDRow
			}{
				Category: &db.AdminGetFAQCategoryWithQuestionCountByIDRow{
					ID:            1,
					Name:          sql.NullString{String: "Test Category", Valid: true},
					Questioncount: 5,
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
	tpl := templates.GetCompiledSiteTemplates(funcs)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tpl.ExecuteTemplate(io.Discard, tc.Name, tc.Data)
			if err != nil {
				t.Errorf("failed to execute template %s: %v", tc.Name, err)
			}
		})
	}
}
