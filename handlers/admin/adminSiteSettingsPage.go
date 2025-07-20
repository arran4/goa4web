package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
)

func AdminSiteSettingsPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		config.AppRuntimeConfig.FeedsEnabled = r.PostFormValue("feeds_enabled") != ""
		langID, _ := strconv.Atoi(r.PostFormValue("default_language"))
		langs, _ := cd.Languages()
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
		*common.CoreData
		Languages          []*db.Language
		SelectedLanguageId int32
	}

	data := Data{
		CoreData:           cd,
		SelectedLanguageId: corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage),
	}
	data.CoreData.FeedsEnabled = config.AppRuntimeConfig.FeedsEnabled
	if langs, err := cd.Languages(); err == nil {
		data.Languages = langs
	}

	handlers.TemplateHandler(w, r, "siteSettingsPage.gohtml", data)
}
