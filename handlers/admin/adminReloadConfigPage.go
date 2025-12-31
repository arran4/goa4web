package admin

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func (h *Handlers) AdminReloadConfigPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Reload Config"
	if cd == nil || !cd.HasRole("administrator") {
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

	cfgMap, err := config.LoadAppConfigFile(core.OSFS{}, h.ConfigFile)
	if err != nil && !errors.Is(err, config.ErrConfigFileNotFound) {
		log.Printf("load config file: %v", err)
	}
	h.Srv.Config = config.NewRuntimeConfig(
		config.WithFileValues(cfgMap),
		config.WithGetenv(os.Getenv),
	)
	if h.DBPool != nil {
		if err := corelanguage.ValidateDefaultLanguage(r.Context(), db.New(h.DBPool), h.Srv.Config.DefaultLanguage); err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
	}

	data.Messages = append(data.Messages, "Configuration reloaded")

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
