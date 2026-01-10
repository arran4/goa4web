package handlers

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

type Page string

func (p Page) Handle(w http.ResponseWriter, r *http.Request, data any) error {
	if err := TemplateHandler(w, r, string(p), data); err != nil {
		return fmt.Errorf("page %s: %w", p, err)
	}
	return nil
}

func (p Page) Exists(opts ...templates.Option) bool {
	return templates.TemplateExists(string(p), opts...)
}

func (p Page) Handler(data any) http.Handler {
	return &templateWithDataHandler{tmpl: string(p), data: data}
}

const (
	TaskErrorAcknowledgementPageTmpl Page = "taskErrorAcknowledgementPage.gohtml"
	NotFoundPageTmpl                 Page = "notFoundPage.gohtml"
	AccessDeniedLoginPageTmpl        Page = "accessDeniedLoginPage.gohtml"
	TaskDoneAutoRefreshPageTmpl      Page = "taskDoneAutoRefreshPage.gohtml"
)
