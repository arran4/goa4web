package privateforum

import (
	"github.com/arran4/goa4web/core/templates"
	"testing"
)

func TestTemplatesExist(t *testing.T) {
	pageTemplates := []string{
		PrivateForumStartDiscussionPageTmpl,
		PrivateForumCreateTopicPageTmpl,
		PrivateForumTopicsOnlyTmpl,
	}

	templates.SetDir("../../core/templates")
	for _, tmpl := range pageTemplates {
		if !templates.IsTemplateAvailable(tmpl) {
			t.Errorf("Template %s not found", tmpl)
		}
	}
}
