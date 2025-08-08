package faq

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "FAQ"

	if _, err := cd.AllAnsweredFAQ(); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	handlers.TemplateHandler(w, r, "faqPage", cd)
}

func CustomFAQIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}
	if data.HasGrant("faq", "question", "post", 0) {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Ask",
			Link: "/faq/ask",
		})
	}
}
