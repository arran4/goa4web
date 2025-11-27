//go:build !linux

package admin

func getSystemStats() (uint64, uint64) {
	return 0, 0
}
