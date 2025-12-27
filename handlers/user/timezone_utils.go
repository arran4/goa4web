package user

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	availableTimezones []string
	tzOnce             sync.Once
)

func getAvailableTimezones() []string {
	tzOnce.Do(func() {
		zoneDir := "/usr/share/zoneinfo"
		_ = filepath.Walk(zoneDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(zoneDir, path)
			if err != nil {
				return err
			}
			if strings.HasPrefix(rel, "posix") || strings.HasPrefix(rel, "right") || strings.HasSuffix(rel, ".tab") {
				return nil
			}
			// LoadLocation to verify it is a valid timezone
			if _, err := time.LoadLocation(rel); err == nil {
				availableTimezones = append(availableTimezones, rel)
			}
			return nil
		})
		sort.Strings(availableTimezones)
	})
	return availableTimezones
}
