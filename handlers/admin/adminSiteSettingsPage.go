package admin

import (
	"net/http"
	"os"
	"sort"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminSiteSettingsPage struct {
	ConfigFile string
}

func (p *AdminSiteSettingsPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminSiteSettingsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Site Settings"
	cd.FeedsEnabled = cd.Config.FeedsEnabled

	values := config.ValuesMap(*cd.Config)
	defaults := config.DefaultMap(config.NewRuntimeConfig())
	usages := config.UsageMap()
	examples := config.ExamplesMap()
	flags := config.NameMap()

	fileVals, _ := config.LoadAppConfigFile(core.OSFS{}, p.ConfigFile)
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
		ConfigFile: p.ConfigFile,
		Config:     cfg,
	}

	AdminSiteSettingsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminSiteSettingsPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Site Settings", "/admin/settings", &AdminPage{}
}

func (p *AdminSiteSettingsPage) PageTitle() string {
	return "Site Settings"
}

var _ tasks.Page = (*AdminSiteSettingsPage)(nil)
var _ tasks.Task = (*AdminSiteSettingsPage)(nil)
var _ http.Handler = (*AdminSiteSettingsPage)(nil)

const AdminSiteSettingsPageTmpl tasks.Template = "admin/siteSettingsPage.gohtml"
