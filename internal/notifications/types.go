package notifications

// SubscriptionTarget exposes a subscribeable object.
type SubscriptionTarget interface {
	// SubscriptionTarget returns the item type and id used when building
	// subscriptions and notifications.
	SubscriptionTarget() (string, int32)
}

// Target references a specific item for subscription notifications.
type Target struct {
	Type string
	ID   int32
}

// SubscriptionTarget implements SubscriptionTarget.
func (t Target) SubscriptionTarget() (string, int32) { return t.Type, t.ID }

// GrantRequirement describes a single permission check.
type GrantRequirement struct {
	Section string
	Item    string
	ItemID  int32
	Action  string
}
