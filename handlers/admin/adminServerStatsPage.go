package admin

import (
	"net/http"
	"runtime"

	common "github.com/arran4/goa4web/handlers/common"
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
		*CoreData
		Stats Stats
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
		Stats: Stats{
			Goroutines: runtime.NumGoroutine(),
			Alloc:      mem.Alloc,
			TotalAlloc: mem.TotalAlloc,
			Sys:        mem.Sys,
			HeapAlloc:  mem.HeapAlloc,
			HeapSys:    mem.HeapSys,
			NumGC:      mem.NumGC,
		},
	}

	common.TemplateHandler(w, r, "serverStatsPage.gohtml", data)
}
