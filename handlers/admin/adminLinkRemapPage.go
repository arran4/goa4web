package admin

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminLinkRemapPage struct{}

func (p *AdminLinkRemapPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		CSV string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Link Remap"
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
		re := regexp.MustCompile(`https?://[^\s"']+`)
		for _, row := range rows {
			if row.News.Valid {
				matches := re.FindAllString(row.News.String, -1)
				for _, m := range matches {
					_ = wcsv.Write([]string{fmt.Sprintf("site_news:%d", row.Idsitenews), m, ""})
				}
			}
		}
		wcsv.Flush()
		data.CSV = buf.String()
	}
	AdminLinkRemapPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminLinkRemapPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Link Remap", "/admin/link-remap", &AdminPage{}
}

func (p *AdminLinkRemapPage) PageTitle() string {
	return "Link Remap"
}

var _ common.Page = (*AdminLinkRemapPage)(nil)
var _ http.Handler = (*AdminLinkRemapPage)(nil)

const AdminLinkRemapPageTmpl tasks.Template = "admin/linkRemapPage.gohtml"
