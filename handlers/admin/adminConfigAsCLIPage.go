package admin

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/configformat"
	"github.com/arran4/goa4web/internal/tasks"
)

// configAsCLIFormat describes a configuration output format for the admin page.
type configAsCLIFormat struct {
	ID          string
	Label       string
	Description string
	Command     string
}

type AdminConfigAsCLIPage struct {
	ConfigFile string
}

func (p *AdminConfigAsCLIPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Config Export"
	cd.FeedsEnabled = cd.Config.FeedsEnabled

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "cli"
	}
	switch format {
	case "cli", "dotenv", "json":
	default:
		format = "cli"
	}

	args := buildConfigAsArgs(r)
	fs := flag.NewFlagSet("as-cli", flag.ContinueOnError)
	opts, err := configformat.ParseAsFlags(fs, args)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse flags: %v", err), http.StatusBadRequest)
		return
	}

	var output string
	switch format {
	case "dotenv":
		output, err = configformat.FormatAsEnvFile(cd.Config, p.ConfigFile, cd.DBRegistry(), opts)
	case "json":
		output, err = configformat.FormatAsJSON(cd.Config, p.ConfigFile)
	default:
		output, err = configformat.FormatAsCLI(cd.Config, p.ConfigFile)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("format output: %v", err), http.StatusInternalServerError)
		return
	}

	formats := []configAsCLIFormat{
		{
			ID:          "cli",
			Label:       "Shell args",
			Description: "Flags to pass to goa4web.",
			Command:     "goa4web config as-cli",
		},
		{
			ID:          "dotenv",
			Label:       "Dotenv",
			Description: "Environment file format with comments.",
			Command:     commandWithExtended("goa4web config as-env-file", opts.Extended),
		},
		{
			ID:          "json",
			Label:       "JSON",
			Description: "Structured configuration output.",
			Command:     "goa4web config as-json",
		},
	}

	data := struct {
		Formats  []configAsCLIFormat
		Format   string
		Output   string
		Extended bool
	}{
		Formats:  formats,
		Format:   format,
		Output:   output,
		Extended: opts.Extended,
	}

	AdminConfigAsCLIPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminConfigAsCLIPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Config Export", "/admin/config/as-cli", &AdminPage{}
}

func (p *AdminConfigAsCLIPage) PageTitle() string {
	return "Config Export"
}

var _ common.Page = (*AdminConfigAsCLIPage)(nil)
var _ http.Handler = (*AdminConfigAsCLIPage)(nil)

func buildConfigAsArgs(r *http.Request) []string {
	if values, ok := r.URL.Query()["extended"]; ok && len(values) > 0 {
		val := strings.TrimSpace(values[0])
		if val == "" || val == "on" || val == "true" || val == "1" {
			return []string{"--extended"}
		}
	}
	return nil
}

func commandWithExtended(command string, extended bool) string {
	if extended {
		return command + " --extended"
	}
	return command
}

// AdminConfigAsCLIPageTmpl renders the admin config export page.
const AdminConfigAsCLIPageTmpl tasks.Template = "admin/configAsCLIPage.gohtml"
