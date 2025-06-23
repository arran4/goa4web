package main

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/config"
)

func adminSiteSettingsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		appRuntimeConfig.FeedsEnabled = r.PostFormValue("feeds_enabled") != ""
		if lang := r.PostFormValue("default_language"); lang != "" {
			tmp := appRuntimeConfig
			tmp.DefaultLanguage = lang
			if err := validateDefaultLanguage(r.Context(), srv.DB, &tmp); err == nil {
				appRuntimeConfig.DefaultLanguage = lang
				if configFile != "" {
					_ = updateConfigKey(configFile, config.EnvDefaultLanguage, lang)
				}
			} else {
				log.Printf("validate default language: %v", err)
			}
		}
		if configFile != "" {
			_ = updateConfigKey(configFile, config.EnvFeedsEnabled, r.PostFormValue("feeds_enabled"))
		}
		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
		return
	}

	type Data struct {
		*CoreData
		Languages       []*Language
		DefaultLanguage string
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	data.CoreData.FeedsEnabled = appRuntimeConfig.FeedsEnabled
	data.DefaultLanguage = appRuntimeConfig.DefaultLanguage
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	if rows, err := queries.FetchLanguages(r.Context()); err == nil {
		data.Languages = rows
	}

	if err := renderTemplate(w, r, "adminSiteSettingsPage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
