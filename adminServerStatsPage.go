package main

import (
	"log"
	"net/http"
	"runtime"
)

func adminServerStatsPage(w http.ResponseWriter, r *http.Request) {
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
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
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

	if err := renderTemplate(w, r, "adminServerStatsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
