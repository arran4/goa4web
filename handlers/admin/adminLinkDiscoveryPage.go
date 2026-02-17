package admin

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

var (
	linkDiscoveryRegex *regexp.Regexp
	linkDiscoveryOnce  sync.Once
)

// AdminLinkDiscoveryPage displays site news URLs for discovery and remapping.
func AdminLinkDiscoveryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		CSV string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Link Discovery & Remap"
	data := Data{}

	if r.URL.Query().Has("generate") {
		q := cd.Queries()
		rows, err := q.GetAllSiteNewsForIndex(r.Context())
		if err != nil {
			log.Printf("list news: %v", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
		var buf bytes.Buffer
		wcsv := csv.NewWriter(&buf)
		_ = wcsv.Write([]string{"internal reference", "original url", "to url"})

		linkDiscoveryOnce.Do(func() {
			linkDiscoveryRegex = regexp.MustCompile(`https?://[^\s"']+`)
		})

		for _, row := range rows {
			if row.News.Valid {
				matches := linkDiscoveryRegex.FindAllString(row.News.String, -1)
				for _, m := range matches {
					_ = wcsv.Write([]string{fmt.Sprintf("site_news:%d", row.Idsitenews), m, ""})
				}
			}
		}
		wcsv.Flush()
		data.CSV = buf.String()
	}
	AdminLinkDiscoveryPageTmpl.Handle(w, r, data)
}

const AdminLinkDiscoveryPageTmpl tasks.Template = "admin/linkDiscoveryPage.gohtml"
