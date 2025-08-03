package user

// subscriptionOption describes a subscription choice presented to the user.
type subscriptionOption struct {
	Name    string
	Pattern string
	Path    string
}

// userSubscriptionOptions lists the available subscription options.
var userSubscriptionOptions = []subscriptionOption{
	{Name: "New blog posts", Pattern: "post:/blog/*", Path: "blogs"},
	{Name: "New articles", Pattern: "post:/writing/*", Path: "writings"},
	{Name: "New news posts", Pattern: "post:/news/*", Path: "news"},
	{Name: "New image board posts", Pattern: "post:/image/*", Path: "images"},
}
