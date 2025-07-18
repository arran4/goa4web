package admin

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func AdminReloadConfigPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Back:     "/admin",
	}

	cfgMap, err := config.LoadAppConfigFile(core.OSFS{}, ConfigFile)
	if err != nil && !errors.Is(err, config.ErrConfigFileNotFound) {
		log.Printf("load config file: %v", err)
	}
	Srv.Config = config.GenerateRuntimeConfig(nil, cfgMap, os.Getenv)
	if err := corelanguage.ValidateDefaultLanguage(r.Context(), db.New(DBPool), Srv.Config.DefaultLanguage); err != nil {
		data.Errors = append(data.Errors, err.Error())
	}

	data.Messages = append(data.Messages, "Configuration reloaded")

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
