package privateforum

import (
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/tasks"
)

type privateForumTask struct {
}

const (
	CreateTopicTmpl handlers.Page = "forum/create_topic.gohtml"
	TopicsOnlyTmpl  handlers.Page = "privateforum/topics_only.gohtml"
)

func NewPrivateForumTask() tasks.Task {
	return &privateForumTask{}
}

func (t *privateForumTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{CreateTopicTmpl, TopicsOnlyTmpl}
}

func (t *privateForumTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *privateForumTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	cd.PageTitle = "Private Forum"
	img, err := share.MakeImageURL(cd.AbsoluteURL(), "Private Forum", "Private discussion forums", cd.ShareSignKey, false)
	if err != nil {
		log.Printf("Error making image URL: %v", err)
	}
	cd.OpenGraph = &common.OpenGraph{
		Title:       "Private Forum",
		Description: "Private discussion forums",
		Image:       img,
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.RequestURI()),
		Type:        "website",
	}

	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		SharedPreviewLoginPageTmpl.Handle(w, r, struct {
			RedirectURL string
		}{
			RedirectURL: url.QueryEscape(r.URL.RequestURI()),
		})
		return
	}
	// Show topics only on the main private page (no creation form)
	TopicsOnlyTmpl.Handle(w, r, nil)
}
