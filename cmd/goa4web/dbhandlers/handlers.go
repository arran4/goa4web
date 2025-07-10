package dbhandlers

import (
	"sync"

	"github.com/arran4/goa4web/config"
)

// BackupHandler backs up a database to the provided file path.
type BackupHandler interface {
	Backup(cfg config.RuntimeConfig, file string) error
}

// RestoreHandler restores a database from the provided file path.
type RestoreHandler interface {
	Restore(cfg config.RuntimeConfig, file string) error
}

var (
	mu              sync.RWMutex
	backupRegistry  = map[string]BackupHandler{}
	restoreRegistry = map[string]RestoreHandler{}
)

// RegisterBackup registers a handler for the given driver name.
func RegisterBackup(driver string, h BackupHandler) {
	mu.Lock()
	backupRegistry[driver] = h
	mu.Unlock()
}

// RegisterRestore registers a handler for the given driver name.
func RegisterRestore(driver string, h RestoreHandler) {
	mu.Lock()
	restoreRegistry[driver] = h
	mu.Unlock()
}

// BackupFor returns the BackupHandler for driver or nil if none registered.
func BackupFor(driver string) BackupHandler {
	mu.RLock()
	h := backupRegistry[driver]
	mu.RUnlock()
	return h
}

// RestoreFor returns the RestoreHandler for driver or nil if none registered.
func RestoreFor(driver string) RestoreHandler {
	mu.RLock()
	h := restoreRegistry[driver]
	mu.RUnlock()
	return h
}

// reset clears all registries. Used by tests.
// Reset clears the internal registries. Used in tests.
func Reset() {
	mu.Lock()
	backupRegistry = map[string]BackupHandler{}
	restoreRegistry = map[string]RestoreHandler{}
	mu.Unlock()
}
