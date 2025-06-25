package goa4web

import (
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"os"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/runtimeconfig"
)

func adminReloadConfigPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
		Back:     "/admin",
	}

	cfgMap := LoadAppConfigFile(ConfigFile)
	srv.Config = runtimeconfig.GenerateRuntimeConfig(nil, cfgMap, os.Getenv)
	if err := corelanguage.ValidateDefaultLanguage(r.Context(), New(dbPool), srv.Config.DefaultLanguage); err != nil {
		data.Errors = append(data.Errors, err.Error())
	}

	data.Messages = append(data.Messages, "Configuration reloaded")

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
