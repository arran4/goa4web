package stats

import (
	"time"
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
