package dbhandlers_test

import (
	"testing"

	"github.com/arran4/goa4web/cmd/goa4web/dbhandlers"
	_ "github.com/arran4/goa4web/cmd/goa4web/dbhandlers/dbdefaults"
)

func TestDefaultRegistrations(t *testing.T) {
	// reset after test to avoid affecting others
	t.Cleanup(dbhandlers.Reset)
	if dbhandlers.BackupFor("mysql") == nil {
		t.Error("mysql backup handler not registered")
	}
	if dbhandlers.RestoreFor("postgres") == nil {
		t.Error("postgres restore handler not registered")
	}
	if dbhandlers.BackupFor("sqlite3") == nil {
		t.Error("sqlite backup handler not registered")
	}
}
