package stats

import (
	"time"

	"github.com/arran4/goa4web/internal/db"
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

// UsageStatsData holds the usage statistics.
type UsageStatsData struct {
	Errors            []string
	ForumTopics       []*db.AdminForumTopicThreadCountsRow
	ForumHandlers     []*db.AdminForumHandlerThreadCountsRow
	ForumCategories   []*db.AdminForumCategoryThreadCountsRow
	WritingCategories []*db.AdminWritingCategoryCountsRow
	LinkerCategories  []*db.GetLinkerCategoryLinkCountsRow
	Imageboards       []*db.AdminImageboardPostCountsRow
	Users             []*db.AdminUserPostCountsRow
	Monthly           []*db.MonthlyUsageRow
	UserMonthly       []*db.UserMonthlyUsageRow
	StartYear         int
}
