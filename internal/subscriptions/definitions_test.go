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
