package faq

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/handlers/share"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "FAQ"
	cd.OpenGraph = &common.OpenGraph{
		Title:       "FAQ",
		Description: "Frequently Asked Questions",
		Image:       share.MakeImageURL(cd.AbsoluteURL(), "FAQ", cd.ShareSigner, false),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
		Type:        "website",
	}

	if _, err := cd.AllAnsweredFAQ(); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	FaqPageTmpl.Handle(w, r, cd)
}

const FaqPageTmpl handlers.Page = "faq/page.gohtml"

func CustomFAQIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}
	if data.HasGrant("faq", "question", "post", 0) {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Ask",
			Link: "/faq/ask",
		})
	}
}
