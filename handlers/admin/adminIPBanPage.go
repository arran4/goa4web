package admin

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminIPBanPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Bans []*db.BannedIp
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "IP Bans"
	data := Data{CoreData: cd}
	queries := cd.Queries()
	rows, err := queries.ListBannedIps(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list banned ips: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Bans = rows
	handlers.TemplateHandler(w, r, "ipBanPage.gohtml", data)
}
