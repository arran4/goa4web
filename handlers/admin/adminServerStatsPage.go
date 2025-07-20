package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"runtime"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
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
		Stats Stats
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
	}

	handlers.TemplateHandler(w, r, "serverStatsPage.gohtml", data)
}
