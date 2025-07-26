package admin

import (
	"net/http"
	"runtime"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
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

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		CoreData: cd,
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
		Config: cd.Config,
	}

	for _, t := range cd.TasksReg.Registered() {
		data.Registries.Tasks = append(data.Registries.Tasks, t.Name())
	}
	if reg := data.CoreData.DBRegistry(); reg != nil {
		data.Registries.DBDrivers = reg.Names()
	}
	if Srv != nil && Srv.DLQReg != nil {
		data.Registries.DLQProviders = Srv.DLQReg.ProviderNames()
	}
	if Srv != nil && Srv.EmailReg != nil {
		data.Registries.EmailProviders = Srv.EmailReg.ProviderNames()
	}
	data.Registries.UploadProviders = upload.ProviderNames()

	handlers.TemplateHandler(w, r, "serverStatsPage.gohtml", data)
}
