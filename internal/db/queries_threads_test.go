package db

import (
	"strings"
	"testing"
)

func TestGetThreadLastPosterAndPerms_AllowsGlobalGrants(t *testing.T) {
	if !strings.Contains(getThreadLastPosterAndPerms, "g.item='topic' OR g.item IS NULL") {
		t.Errorf("missing global item check")
	}
	if !strings.Contains(getThreadLastPosterAndPerms, "g.item_id = t.idforumtopic OR g.item_id IS NULL") {
		t.Errorf("missing global item_id check")
	}
}
