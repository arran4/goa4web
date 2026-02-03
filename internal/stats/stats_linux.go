//go:build linux

package stats

import "golang.org/x/sys/unix"

func getSystemStats() (uint64, uint64) {
	var diskFree uint64
	if varStat := new(unix.Statfs_t); unix.Statfs("/", varStat) == nil {
		diskFree = varStat.Bavail * uint64(varStat.Bsize)
	}

	var sysInfo unix.Sysinfo_t
	var ramFree uint64
	if unix.Sysinfo(&sysInfo) == nil {
		unit := uint64(sysInfo.Unit)
		if unit == 0 {
			unit = 1
		}
		ramFree = sysInfo.Freeram * unit
	}
	return diskFree, ramFree
}
