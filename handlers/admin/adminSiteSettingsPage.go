package admin

import (
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
)

func AdminSiteSettingsPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cd := r.Context().Value(common.KeyCoreData).(*CoreData)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		config.AppRuntimeConfig.FeedsEnabled = r.PostFormValue("feeds_enabled") != ""
		langID, _ := strconv.Atoi(r.PostFormValue("default_language"))
		langs, _ := cd.AllLanguages()
		name := ""
		for _, l := range langs {
			if int(l.Idlanguage) == langID {
				name = l.Nameof.String
				break
			}
		}
		config.AppRuntimeConfig.DefaultLanguage = name
		if err := updateConfigKey(ConfigFile, config.EnvDefaultLanguage, name); err != nil {
			log.Printf("config write error: %v", err)
		}
		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
		return
	}

	type Data struct {
		*CoreData
		Languages          []*db.Language
		SelectedLanguageId int32
	}

	data := Data{
		CoreData:           cd,
		SelectedLanguageId: corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage),
	}
	data.CoreData.FeedsEnabled = config.AppRuntimeConfig.FeedsEnabled
	if langs, err := cd.AllLanguages(); err == nil {
		data.Languages = langs
	}

	if err := templates.RenderTemplate(w, "siteSettingsPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
