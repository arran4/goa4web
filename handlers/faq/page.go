package faq

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/handlers/share"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "FAQ"
	imageURL, _ := share.MakeImageURL(cd.AbsoluteURL(), "FAQ", "Frequently Asked Questions", cd.ShareSignKey, false)
	cd.OpenGraph = &common.OpenGraph{
		Title:       "FAQ",
		Description: "Frequently Asked Questions",
		Image:       imageURL,
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
		Type:        "website",
	}

	if _, err := cd.AllAnsweredFAQ(); err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	FaqPageTmpl.Handle(w, r, cd)
}

const FaqPageTmpl tasks.Template = "faq/page.gohtml"

func CustomFAQIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}
	if data.HasGrant("faq", "question", "post", 0) {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Ask",
			Link: "/faq/ask",
		})
	}
}
