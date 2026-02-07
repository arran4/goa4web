package admin

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/configexplain"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminConfigExplainPage struct {
	ConfigFile string
}

func (p *AdminConfigExplainPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Config Explain"

	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, p.ConfigFile)
	if err != nil && !errors.Is(err, config.ErrConfigFileNotFound) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("load config file: %w", err))
		return
	}

	values := config.ValuesMap(*cd.Config)
	infos := configexplain.Explain(configexplain.Inputs{
		FileValues: fileVals,
		ConfigFile: p.ConfigFile,
		Values:     values,
		Getenv:     os.Getenv,
	})

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

	filtered := make([]configexplain.OptionInfo, 0, len(infos))
	for _, info := range infos {
		if _, ok := hide[info.Env]; ok {
			continue
		}
		filtered = append(filtered, info)
	}

	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Env < filtered[j].Env })

	data := struct {
		ConfigFile string
		Options    []configexplain.OptionInfo
	}{
		ConfigFile: p.ConfigFile,
		Options:    filtered,
	}

	AdminConfigExplainPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminConfigExplainPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Config Explain", "/admin/config/explain", &AdminPage{}
}

func (p *AdminConfigExplainPage) PageTitle() string {
	return "Config Explain"
}

var _ common.Page = (*AdminConfigExplainPage)(nil)
var _ http.Handler = (*AdminConfigExplainPage)(nil)

// AdminConfigExplainPageTmpl renders the configuration explain admin page.
const AdminConfigExplainPageTmpl tasks.Template = "admin/configExplainPage.gohtml"
