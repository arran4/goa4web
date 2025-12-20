package subscriptions

type Definition struct {
	Name        string
	Description string
	Pattern     string
	IsAdmin     bool
}

var Definitions = []Definition{
	{
		Name:        "Subscribe to all threads",
		Description: "Subscribe to any thread created on any topic",
		Pattern:     "subscribe:/forum/topic/*/thread/*",
	},
	{
		Name:        "Notify on all threads",
		Description: "Notify to any thread created on any topic",
		Pattern:     "notify:/forum/topic/*/thread/*",
	},
	{
		Name:        "Subscribe to topic threads",
		Description: "Subscribe to any thread created on a particular topic",
		Pattern:     "subscribe:/forum/topic/{topicid}/thread/*",
	},
	{
		Name:        "Notify on topic threads",
		Description: "Notify to any thread created on a particular topic",
		Pattern:     "notify:/forum/topic/{topicid}/thread/*",
	},
}
