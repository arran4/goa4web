package privateforum

import (
	"testing"

	coretemplates "github.com/arran4/goa4web/core/templates"
)

func TestTemplatesExist(t *testing.T) {
	pageTemplates := []string{
		PrivateForumStartDiscussionPageTmpl,
		PrivateForumCreateTopicPageTmpl,
		PrivateForumTopicsOnlyTmpl,
	}

	dir := "../../core/templates"
	for _, tmpl := range pageTemplates {
		if !coretemplates.IsTemplateAvailable(tmpl, dir) {
			t.Errorf("Template %s not found", tmpl)
		}
	}
}
