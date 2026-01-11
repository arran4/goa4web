package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminIPBanPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Bans []*db.BannedIp
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "IP Bans"
	data := Data{}
	queries := cd.Queries()
	rows, err := queries.ListBannedIps(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list banned ips: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Bans = rows
	AdminIPBanPageTmpl.Handle(w, r, data)
}

const AdminIPBanPageTmpl handlers.Page = "ipBanPage.gohtml"
