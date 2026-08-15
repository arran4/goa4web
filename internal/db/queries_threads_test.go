package db

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/consts"
)

func TestGetThreadLastPosterAndPerms_AllowsGlobalGrants(t *testing.T) {
	if !strings.Contains(getThreadLastPosterAndPerms, "g.item='topic' OR g.item IS NULL") {
		t.Errorf("missing global item check")
	}
	if !strings.Contains(getThreadLastPosterAndPerms, "g.item_id = t.idforumtopic OR g.item_id IS NULL") {
		t.Errorf("missing global item_id check")
	}
}

func TestGetThreadLastPosterAndPerms_UsesPrivateThreadGrants(t *testing.T) {
	want := fmt.Sprintf("g.section='%s'", consts.PermissionSectionPrivateForumThread)
	if !strings.Contains(getThreadLastPosterAndPerms, want) {
		t.Errorf("private thread access does not use privateforum_thread grants")
	}
}
