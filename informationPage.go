package main

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"log"
	"net/http"
	"time"
)

func informationPage(w http.ResponseWriter, r *http.Request) {

	type SystemInformation struct {
		Processors  []cpu.InfoStat
		Uptime      time.Duration
		IdlePercent float64
		LoadAverage *load.AvgStat
	}

	type Data struct {
		*CoreData
		System *SystemInformation
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	ld, err := load.Avg()
	if err != nil {
		log.Printf("load.Avg Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	uptime, err := host.UptimeWithContext(r.Context())
	if err != nil {
		log.Printf("load.Avg Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ts, err := cpu.Times(false)
	if err != nil {
		log.Printf("cpu.Times Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.System = &SystemInformation{
		LoadAverage: ld,
		Uptime:      time.Second * time.Duration(uptime),
		IdlePercent: ts[0].Idle * 100 / ts[0].System,
	}
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Printf("cpu.Times Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.System.Processors = cpuInfo

	renderTemplate(w, r, "informationPage.gohtml", data)
}
