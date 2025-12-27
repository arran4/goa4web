package faq

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"net/http"
)

// AdminQuestionsPage renders the questions administration view.
func AdminQuestionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	d := cd.Queries()
	p := &AdminQuestions{
		CoreData: common.CoreData{
			Config: cd.Config,
		},
	}
	if err := p.Load(r.Context(), d.(*db.Queries), r); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	if err := cd.ExecuteSiteTemplate(w, r, p.TemplateName(), p); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
