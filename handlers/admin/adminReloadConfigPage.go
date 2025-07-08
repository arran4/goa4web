package admin

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/runtimeconfig"
)

func AdminReloadConfigPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
		Back:     "/admin",
	}

	cfgMap, err := config.LoadAppConfigFile(core.OSFS{}, ConfigFile)
	if err != nil && !errors.Is(err, config.ErrConfigFileNotFound) {
		log.Printf("load config file: %v", err)
	}
	Srv.Config = runtimeconfig.GenerateRuntimeConfig(nil, cfgMap, os.Getenv)
	if err := corelanguage.ValidateDefaultLanguage(r.Context(), db.New(DBPool), Srv.Config.DefaultLanguage); err != nil {
		data.Errors = append(data.Errors, err.Error())
	}

	data.Messages = append(data.Messages, "Configuration reloaded")

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
