package admin

import (
	"net/http"
	"os"
	"sort"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/config"
)

func (h *Handlers) AdminSiteSettingsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Site Settings"
	cd.FeedsEnabled = cd.Config.FeedsEnabled

	values := config.ValuesMap(*cd.Config)
	defaults := config.DefaultMap(config.NewRuntimeConfig())
	usages := config.UsageMap()
	examples := config.ExamplesMap()
	flags := config.NameMap()

	fileVals, _ := config.LoadAppConfigFile(core.OSFS{}, h.ConfigFile)
	hide := map[string]struct{}{
		config.EnvDBConn:              {},
		config.EnvSMTPPass:            {},
		config.EnvJMAPPass:            {},
		config.EnvSendGridKey:         {},
		config.EnvSessionSecret:       {},
		config.EnvSessionSecretFile:   {},
		config.EnvImageSignSecret:     {},
		config.EnvImageSignSecretFile: {},
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		if _, ok := hide[k]; ok {
			delete(values, k)
			continue
		}
		if values[k] == "" {
			delete(values, k)
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	type detail struct {
		Env     string
		Flag    string
		Value   string
		Default string
		Usage   string
		Example []string
		Source  string
	}
	cfg := make([]detail, 0, len(keys))
	for _, k := range keys {
		src := "default"
		if v := fileVals[k]; v != "" && v == values[k] {
			src = "file"
		} else if v := os.Getenv(k); v != "" && v == values[k] {
			src = "env"
		} else if values[k] != defaults[k] {
			src = "flag"
		}
		cfg = append(cfg, detail{
			Env:     k,
			Flag:    flags[k],
			Value:   values[k],
			Default: defaults[k],
			Usage:   usages[k],
			Example: examples[k],
			Source:  src,
		})
	}

	data := struct {
		ConfigFile string
		Config     []detail
	}{
		ConfigFile: h.ConfigFile,
		Config:     cfg,
	}

	AdminSiteSettingsPageTmpl.Handle(w, r, data)
}

const AdminSiteSettingsPageTmpl handlers.Page = "admin/siteSettingsPage.gohtml"
