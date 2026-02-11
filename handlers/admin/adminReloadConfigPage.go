package admin

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	corelanguage "github.com/arran4/goa4web/core/language"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/internal/db"
)

type AdminReloadConfigPage struct {
	ConfigFile string
	Srv        *server.Server
	DBPool     *sql.DB
}

func (p *AdminReloadConfigPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Reload Config"
	if cd == nil || !cd.HasAdminRole() {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/admin",
	}

	cfgMap, err := config.LoadAppConfigFile(core.OSFS{}, p.ConfigFile)
	if err != nil && !errors.Is(err, config.ErrConfigFileNotFound) {
		log.Printf("load config file: %v", err)
	}
	p.Srv.Config = config.NewRuntimeConfig(
		config.WithFileValues(cfgMap),
		config.WithGetenv(os.Getenv),
	)
	if p.DBPool != nil {
		if err := corelanguage.ValidateDefaultLanguage(r.Context(), db.New(p.DBPool), p.Srv.Config.DefaultLanguage); err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
	}

	data.Messages = append(data.Messages, "Configuration reloaded")

	RunTaskPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminReloadConfigPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Reload Config", "/admin/reload", &AdminPage{}
}

func (p *AdminReloadConfigPage) PageTitle() string {
	return "Reload Config"
}

var _ common.Page = (*AdminReloadConfigPage)(nil)
var _ http.Handler = (*AdminReloadConfigPage)(nil)
