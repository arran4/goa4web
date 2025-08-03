package user

// subscriptionOption defines a user subscription preference for notifications.
type subscriptionOption struct {
	Name    string
	Pattern string
	Path    string
}

var userSubscriptionOptions = []subscriptionOption{
	{Name: "New blog posts", Pattern: "post:/blog/*", Path: "blogs"},
	{Name: "New articles", Pattern: "post:/writing/*", Path: "writings"},
	{Name: "New news posts", Pattern: "post:/news/*", Path: "news"},
	{Name: "New image board posts", Pattern: "post:/image/*", Path: "images"},
}
