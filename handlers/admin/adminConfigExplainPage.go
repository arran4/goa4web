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

// AdminConfigExplainPage shows configuration sources for the current runtime config.
func (h *Handlers) AdminConfigExplainPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Config Explain"

	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, h.ConfigFile)
	if err != nil && !errors.Is(err, config.ErrConfigFileNotFound) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("load config file: %w", err))
		return
	}

	values := config.ValuesMap(*cd.Config)
	infos := configexplain.Explain(configexplain.Inputs{
		FileValues: fileVals,
		ConfigFile: h.ConfigFile,
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
		ConfigFile: h.ConfigFile,
		Options:    filtered,
	}

	AdminConfigExplainPageTmpl.Handle(w, r, data)
}

// AdminConfigExplainPageTmpl renders the configuration explain admin page.
const AdminConfigExplainPageTmpl tasks.Template = "admin/configExplainPage.gohtml"
