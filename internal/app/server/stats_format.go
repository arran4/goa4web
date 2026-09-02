package server

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/arran4/goa4web/internal/stats"
)

var serverStatsTmpl = template.Must(template.New("server").Parse(`
Server Stats: Uptime={{.Uptime}} Goroutines={{.Stats.Goroutines}} Mem(Alloc={{.Stats.Alloc}} Total={{.Stats.TotalAlloc}} Sys={{.Stats.Sys}} HeapAlloc={{.Stats.HeapAlloc}} HeapSys={{.Stats.HeapSys}}) GC={{.Stats.NumGC}} CPU={{.Stats.NumCPU}}({{.Stats.Arch}}) Free(Disk={{.Stats.DiskFree}} RAM={{.Stats.RAMFree}}) AutoSubFailures={{.Stats.AutoSubscribePreferenceFailures}} Reg(Tasks={{len .Registries.Tasks}} DB={{len .Registries.DBDrivers}} DLQ={{len .Registries.DLQProviders}} Email={{len .Registries.EmailProviders}} Upload={{len .Registries.UploadProviders}} Router={{len .Registries.RouterModules}})
`))

var usageStatsTmpl = template.Must(template.New("usage").Parse(`
Usage Stats: Errors={{len .Errors}}{{if .Errors}} [{{range $i, $e := .Errors}}{{if $i}}, {{end}}{{$e}}{{end}}]{{end}} Forum(Topics={{len .ForumTopics}} Handlers={{len .ForumHandlers}} Categories={{len .ForumCategories}}) Categories(Writing={{len .WritingCategories}} Linker={{len .LinkerCategories}}) Imageboards={{len .Imageboards}} Users={{len .Users}} Monthly(Global={{len .Monthly}} User={{len .UserMonthly}}) StartYear={{.StartYear}}
`))

func formatServerStats(data stats.ServerStatsData) (string, error) {
	var buf bytes.Buffer
	err := serverStatsTmpl.Execute(&buf, data)
	return strings.TrimSpace(buf.String()), err
}

func formatUsageStats(data stats.UsageStatsData) (string, error) {
	var buf bytes.Buffer
	err := usageStatsTmpl.Execute(&buf, data)
	return strings.TrimSpace(buf.String()), err
}
