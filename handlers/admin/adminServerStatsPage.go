package admin

import (
	"net/http"
	"runtime"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload"
)

func AdminServerStatsPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Goroutines int
		Alloc      uint64
		TotalAlloc uint64
		Sys        uint64
		HeapAlloc  uint64
		HeapSys    uint64
		NumGC      uint32
	}

	type Data struct {
		*common.CoreData
		Stats      Stats
		Uptime     time.Duration
		Config     config.RuntimeConfig
		Registries struct {
			Tasks           []string
			DBDrivers       []string
			DLQProviders    []string
			EmailProviders  []string
			UploadProviders []string
		}
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Stats: Stats{
			Goroutines: runtime.NumGoroutine(),
			Alloc:      mem.Alloc,
			TotalAlloc: mem.TotalAlloc,
			Sys:        mem.Sys,
			HeapAlloc:  mem.HeapAlloc,
			HeapSys:    mem.HeapSys,
			NumGC:      mem.NumGC,
		},
		Uptime: time.Since(StartTime),
		Config: config.AppRuntimeConfig,
	}

	for _, t := range tasks.Registered() {
		data.Registries.Tasks = append(data.Registries.Tasks, t.Name())
	}
	if reg := data.CoreData.DBRegistry(); reg != nil {
		data.Registries.DBDrivers = reg.Names()
	}
	data.Registries.DLQProviders = dlq.ProviderNames()
	data.Registries.EmailProviders = email.ProviderNames()
	data.Registries.UploadProviders = upload.ProviderNames()

	handlers.TemplateHandler(w, r, "serverStatsPage.gohtml", data)
}
