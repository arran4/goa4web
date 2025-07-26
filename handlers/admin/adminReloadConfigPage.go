package admin

import (
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminReloadConfigPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasRole("administrator") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     "/admin",
	}

	cfgMap, err := config.LoadAppConfigFile(core.OSFS{}, ConfigFile)
	if err != nil && !errors.Is(err, config.ErrConfigFileNotFound) {
		log.Printf("load config file: %v", err)
	}
	Srv.Config = config.NewRuntimeConfig(
		config.WithFileValues(cfgMap),
		config.WithGetenv(os.Getenv),
	)
	if err := corelanguage.ValidateDefaultLanguage(r.Context(), db.New(DBPool), Srv.Config.DefaultLanguage); err != nil {
		data.Errors = append(data.Errors, err.Error())
	}

	data.Messages = append(data.Messages, "Configuration reloaded")

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
