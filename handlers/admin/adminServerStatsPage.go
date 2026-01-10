package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/upload"
)

func (h *Handlers) AdminServerStatsPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Goroutines int
		Alloc      uint64
		TotalAlloc uint64
		Sys        uint64
		HeapAlloc  uint64
		HeapSys    uint64
		NumGC      uint32
		NumCPU     int
		Arch       string
		DiskFree   uint64
		RAMFree    uint64
	}

	type Data struct {
		Stats      Stats
		Uptime     time.Duration
		ConfigEnv  string
		ConfigJSON string
		Registries struct {
			Tasks           []string
			DBDrivers       []string
			DLQProviders    []string
			EmailProviders  []string
			UploadProviders []string
			RouterModules   []string
		}
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	diskFree, ramFree := getSystemStats()

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Server Stats"
	data := Data{
		Stats: Stats{
			Goroutines: runtime.NumGoroutine(),
			Alloc:      mem.Alloc,
			TotalAlloc: mem.TotalAlloc,
			Sys:        mem.Sys,
			HeapAlloc:  mem.HeapAlloc,
			HeapSys:    mem.HeapSys,
			NumGC:      mem.NumGC,
			NumCPU:     runtime.NumCPU(),
			Arch:       runtime.GOARCH,
			DiskFree:   diskFree,
			RAMFree:    ramFree,
		},
		Uptime: time.Since(StartTime),
	}

	envMap, err := config.ToEnvMap(cd.Config, h.ConfigFile)
	if err == nil {
		defMap, _ := config.ToEnvMap(config.NewRuntimeConfig(), "")
		usage := config.UsageMap()
		ext := config.ExtendedUsageMap(cd.DBRegistry())
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

	for _, t := range cd.TasksReg.Registered() {
		data.Registries.Tasks = append(data.Registries.Tasks, t.Name())
	}
	if reg := cd.DBRegistry(); reg != nil {
		data.Registries.DBDrivers = reg.Names()
	}
	if h.Srv != nil {
		if h.Srv.DLQReg != nil {
			data.Registries.DLQProviders = h.Srv.DLQReg.ProviderNames()
		}
		if h.Srv.EmailReg != nil {
			data.Registries.EmailProviders = h.Srv.EmailReg.ProviderNames()
		}
		if h.Srv.RouterReg != nil {
			data.Registries.RouterModules = h.Srv.RouterReg.Names()
		}
	}
	data.Registries.UploadProviders = upload.ProviderNames()

	AdminServerStatsPageTmpl.Handle(w, r, data)
}

const AdminServerStatsPageTmpl handlers.Page = "admin/serverStatsPage.gohtml"
