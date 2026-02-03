package stats

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload"
)

// ServerStatsMetrics holds runtime and system usage metrics.
type ServerStatsMetrics struct {
	Goroutines                      int
	Alloc                           uint64
	TotalAlloc                      uint64
	Sys                             uint64
	HeapAlloc                       uint64
	HeapSys                         uint64
	NumGC                           uint32
	NumCPU                          int
	Arch                            string
	DiskFree                        uint64
	RAMFree                         uint64
	AutoSubscribePreferenceFailures int64
}

// ServerStatsRegistries describes the registered component providers.
type ServerStatsRegistries struct {
	Tasks           []string
	DBDrivers       []string
	DLQProviders    []string
	EmailProviders  []string
	UploadProviders []string
	RouterModules   []string
}

// ServerStatsData bundles stats and registry information for reporting.
type ServerStatsData struct {
	Stats        ServerStatsMetrics
	Uptime       time.Duration
	ConfigEnv    string
	ConfigJSON   string
	ConfigValues map[string]string
	Registries   ServerStatsRegistries
}

// BuildServerStatsData assembles server metrics and configuration details.
func BuildServerStatsData(cfg *config.RuntimeConfig, configFile string, tasksReg *tasks.Registry, dbReg *dbdrivers.Registry, dlqReg *dlq.Registry, emailReg *email.Registry, routerModules []string) ServerStatsData {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	diskFree, ramFree := getSystemStats()

	data := ServerStatsData{
		Stats: ServerStatsMetrics{
			Goroutines:                      runtime.NumGoroutine(),
			Alloc:                           mem.Alloc,
			TotalAlloc:                      mem.TotalAlloc,
			Sys:                             mem.Sys,
			HeapAlloc:                       mem.HeapAlloc,
			HeapSys:                         mem.HeapSys,
			NumGC:                           mem.NumGC,
			NumCPU:                          runtime.NumCPU(),
			Arch:                            runtime.GOARCH,
			DiskFree:                        diskFree,
			RAMFree:                         ramFree,
			AutoSubscribePreferenceFailures: AutoSubscribePreferenceFailures.Load(),
		},
	}

	if !StartTime.IsZero() {
		data.Uptime = time.Since(StartTime)
	}

	if cfg != nil {
		envMap, err := config.ToEnvMap(cfg, configFile)
		if err == nil {
			data.ConfigValues = envMap
			defMap, _ := config.ToEnvMap(config.NewRuntimeConfig(), "")
			usage := config.UsageMap()
			ext := config.ExtendedUsageMap(dbReg)
			ex := config.ExamplesMap()
			keys := make([]string, 0, len(envMap))
			for k := range envMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			var sb strings.Builder
			for _, k := range keys {
				line := "# "
				if u := usage[k]; u != "" {
					line += fmt.Sprintf("%s (default: %s)", u, defMap[k])
				} else {
					line += fmt.Sprintf("default: %s", defMap[k])
				}
				if xs := ex[k]; len(xs) > 0 {
					line += fmt.Sprintf(" (examples: %s)", strings.Join(xs, ", "))
				}
				sb.WriteString(line + "\n")
				if e := ext[k]; e != "" {
					for _, ln := range strings.Split(strings.TrimSuffix(e, "\n"), "\n") {
						sb.WriteString("# " + ln + "\n")
					}
				}
				sb.WriteString(fmt.Sprintf("%s=%s\n", k, envMap[k]))
			}
			data.ConfigEnv = sb.String()
			if b, err := json.MarshalIndent(envMap, "", "  "); err == nil {
				data.ConfigJSON = string(b)
			}
		}
	}

	if tasksReg != nil {
		for _, t := range tasksReg.Registered() {
			data.Registries.Tasks = append(data.Registries.Tasks, t.Name())
		}
	}
	if dbReg != nil {
		data.Registries.DBDrivers = dbReg.Names()
	}
	if dlqReg != nil {
		data.Registries.DLQProviders = dlqReg.ProviderNames()
	}
	if emailReg != nil {
		data.Registries.EmailProviders = emailReg.ProviderNames()
	}
	if routerModules != nil {
		data.Registries.RouterModules = routerModules
	}
	data.Registries.UploadProviders = upload.ProviderNames()

	return data
}
