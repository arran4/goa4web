//go:build !linux

package stats

func getSystemStats() (uint64, uint64) {
	return 0, 0
}
