package handlers

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

type Page = tasks.Template

func init() {
	tasks.Handle = func(w http.ResponseWriter, r *http.Request, p tasks.Template, data any) error {
		return TemplateHandler(w, r, p, data)
	}
	tasks.TemplateExecute = func(w http.ResponseWriter, r *http.Request, p tasks.Template, data any) error {
		cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd == nil {
			return fmt.Errorf("core data not found in context")
		}
		return cd.ExecuteSiteTemplate(w, r, string(p), data)
	}
	tasks.Handler = func(p tasks.Template, data any) http.Handler {
		return &templateWithDataHandler{tmpl: p, data: data}
	}
}

const (
	TaskErrorAcknowledgementPageTmpl Page = "pages/misc/taskErrorAcknowledgementPage.gohtml"
	NotFoundPageTmpl                 Page = "pages/misc/notFoundPage.gohtml"
	AccessDeniedLoginPageTmpl        Page = "pages/auth/accessDeniedLoginPage.gohtml"
	TaskDoneAutoRefreshPageTmpl      Page = "pages/misc/taskDoneAutoRefreshPage.gohtml"
)
