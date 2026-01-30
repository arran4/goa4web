package subscriptions

import (
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestGetUserSubscriptions_UnknownPattern(t *testing.T) {
	// Create a subscription with a pattern that doesn't match any known definition
	dbSubs := []*db.ListSubscriptionsByUserRow{
		{
			ID:      1,
			Pattern: "unknown:pattern:123",
			Method:  "email",
		},
	}

	// This should not panic
	groups := GetUserSubscriptions(dbSubs)

	if len(groups) == 0 {
		t.Errorf("Expected at least one group, got 0")
	}

	found := false
	for _, g := range groups {
		if g.Pattern == "unknown:pattern:123" {
			found = true
			if len(g.Instances) != 1 {
				t.Errorf("Expected 1 instance, got %d", len(g.Instances))
			}
			break
		}
	}

	if !found {
		t.Errorf("Expected to find group with pattern 'unknown:pattern:123'")
	}
}

func TestGetUserSubscriptions_KnownPattern(t *testing.T) {
	// Create a subscription with a known pattern
	dbSubs := []*db.ListSubscriptionsByUserRow{
		{
			ID:      1,
			Pattern: "create thread:/forum/topic/*",
			Method:  "email",
		},
	}

	// This should not panic
	groups := GetUserSubscriptions(dbSubs)

	found := false
	for _, g := range groups {
		if g.Pattern == "create thread:/forum/topic/*" {
			found = true
			if len(g.Instances) != 1 {
				t.Errorf("Expected 1 instance, got %d", len(g.Instances))
			}
			break
		}
	}

	if !found {
		t.Errorf("Expected to find group with pattern 'create thread:/forum/topic/*'")
	}
}

func TestGetUserSubscriptions_ReportedIssues(t *testing.T) {
	dbSubs := []*db.ListSubscriptionsByUserRow{
		{
			ID:      1,
			Pattern: "private topic create:/private",
			Method:  "email",
		},
		{
			ID:      2,
			Pattern: "write reply:/forum/topic/12/thread/12/*",
			Method:  "email",
		},
	}

	groups := GetUserSubscriptions(dbSubs)

	expectedMap := map[string]string{
		"private topic create:/private":           "Private Topic Created",
		"write reply:/forum/topic/12/thread/12/*": "Write Reply (Legacy)",
	}

	for _, g := range groups {
		for _, inst := range g.Instances {
			if expectedName, ok := expectedMap[inst.Original]; ok {
				if g.Definition.Name != expectedName {
					t.Errorf("For pattern '%s', expected definition name '%s', got '%s'", inst.Original, expectedName, g.Definition.Name)
				}
				delete(expectedMap, inst.Original)
			}
		}
	}

	if len(expectedMap) > 0 {
		for k, v := range expectedMap {
			t.Errorf("Pattern '%s' (expected '%s') was not found in any group", k, v)
		}
	}
}

func TestGetUserSubscriptions_LegacyUpgrade(t *testing.T) {
	dbSubs := []*db.ListSubscriptionsByUserRow{
		{
			ID:      1,
			Pattern: "write reply:/forum/topic/123/thread/456/*",
			Method:  "email",
		},
	}

	groups := GetUserSubscriptions(dbSubs)

	found := false
	for _, g := range groups {
		if g.Definition.Name == "Write Reply (Legacy)" {
			found = true
			if !g.Definition.HideIfNone {
				t.Errorf("Expected Write Reply (Legacy) to be hidden if none")
			}
			if len(g.Instances) != 1 {
				t.Errorf("Expected 1 instance, got %d", len(g.Instances))
			} else {
				inst := g.Instances[0]
				expectedUpgrade := "reply:/forum/topic/123/thread/456/*"
				if inst.UpgradeTo != expectedUpgrade {
					t.Errorf("Expected UpgradeTo '%s', got '%s'", expectedUpgrade, inst.UpgradeTo)
				}
			}
		}
	}

	if !found {
		t.Errorf("Expected to find 'Write Reply (Legacy)' group")
	}
}
