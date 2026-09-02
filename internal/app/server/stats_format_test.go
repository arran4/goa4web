package server

import (
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/stats"
)

func TestFormatServerStats(t *testing.T) {
	data := stats.ServerStatsData{
		Uptime: 5 * time.Minute,
		Stats: stats.ServerStatsMetrics{
			Goroutines:                      12,
			Alloc:                           1234,
			TotalAlloc:                      12345,
			Sys:                             54321,
			HeapAlloc:                       1234,
			HeapSys:                         12345,
			NumGC:                           2,
			NumCPU:                          8,
			Arch:                            "amd64",
			DiskFree:                        123456789,
			RAMFree:                         987654321,
			AutoSubscribePreferenceFailures: 3,
		},
		Registries: stats.ServerStatsRegistries{
			Tasks:           []string{"Task1", "Task2"},
			DBDrivers:       []string{"mysql"},
			DLQProviders:    []string{"dir"},
			EmailProviders:  []string{"smtp", "local"},
			UploadProviders: []string{"local"},
			RouterModules:   []string{"auth", "forum"},
		},
	}

	out, err := formatServerStats(data)
	if err != nil {
		t.Fatalf("formatServerStats error: %v", err)
	}

	expectedParts := []string{
		"Uptime=5m0s",
		"Goroutines=12",
		"Alloc=1234",
		"Total=12345",
		"Sys=54321",
		"HeapAlloc=1234",
		"HeapSys=12345",
		"GC=2",
		"CPU=8(amd64)",
		"Disk=123456789",
		"RAM=987654321",
		"AutoSubFailures=3",
		"Tasks=[Task1,Task2]",
		"DB=[mysql]",
		"DLQ=[dir]",
		"Email=[smtp,local]",
		"Upload=[local]",
		"Router=[auth,forum]",
	}

	for _, part := range expectedParts {
		if !strings.Contains(out, part) {
			t.Errorf("output missing expected part %q. Got:\n%s", part, out)
		}
	}
}

func TestFormatUsageStats(t *testing.T) {
	data := stats.UsageStatsData{
		Errors: []string{"query error 1", "timeout error"},
		ForumTopics: []*db.AdminForumTopicThreadCountsRow{
			{}, {}, {},
		},
		ForumHandlers: []*db.AdminForumHandlerThreadCountsRow{
			{}, {},
		},
		ForumCategories: []*db.AdminForumCategoryThreadCountsRow{
			{},
		},
		WritingCategories: []*db.AdminWritingCategoryCountsRow{
			{},
		},
		LinkerCategories: []*db.GetLinkerCategoryLinkCountsRow{
			{},
		},
		Imageboards: []*db.AdminImageboardPostCountsRow{
			{},
		},
		Users: []*db.AdminUserPostCountsRow{
			{}, {}, {}, {},
		},
		Monthly: []*db.MonthlyUsageRow{
			{},
		},
		UserMonthly: []*db.UserMonthlyUsageRow{
			{},
		},
		StartYear: 2005,
	}

	out, err := formatUsageStats(data)
	if err != nil {
		t.Fatalf("formatUsageStats error: %v", err)
	}

	expectedParts := []string{
		"Errors=2 [query error 1, timeout error]",
		"Topics=3",
		"Handlers=2",
		"Categories=1",
		"Writing=1",
		"Linker=1",
		"Imageboards=1",
		"Users=4",
		"Global=1",
		"User=1",
		"StartYear=2005",
	}

	for _, part := range expectedParts {
		if !strings.Contains(out, part) {
			t.Errorf("output missing expected part %q. Got:\n%s", part, out)
		}
	}
}
