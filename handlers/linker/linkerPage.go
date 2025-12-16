package linker

import (
	"github.com/arran4/goa4web/core/common"
	"net/http"
	"strings"
)

func LinkerPage(w http.ResponseWriter, r *http.Request) {
	t := NewLinkerTask().(*linkerTask)
	t.Get(w, r)
}

func CustomLinkerIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}
	if r.URL.Path == "/linker" || strings.HasPrefix(r.URL.Path, "/linker/category/") {
		data.RSSFeedURL = "/linker/rss"
		data.AtomFeedURL = "/linker/atom"
	}
}
