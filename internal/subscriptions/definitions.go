package subscriptions

import (
	"strings"
)

type Definition struct {
	Name        string
	Description string
	Method      string
	Pattern     string
	IsAdmin     bool
}

// Action returns the action part of the pattern (e.g., "subscribe", "notify", "reply").
func (d Definition) Action() string {
	parts := strings.SplitN(d.Pattern, ":", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// URLPattern returns the URL part of the pattern.
func (d Definition) URLPattern() string {
	parts := strings.SplitN(d.Pattern, ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

var Definitions = []Definition{
	// Forum
	{
		Name:        "New Threads (All)",
		Description: "Notify when a new thread is created in any topic",
		Method:      "internal",
		Pattern:     "create thread:/forum/topic/*",
	},
	{
		Name:        "New Threads (Specific Topic)",
		Description: "Notify when a new thread is created in this topic",
		Method:      "internal",
		Pattern:     "create thread:/forum/topic/{topicid}/*",
	},
	{
		Name:        "Replies (All)",
		Description: "Notify when a reply is posted in any thread",
		Method:      "internal",
		Pattern:     "reply:/forum/topic/*/thread/*",
	},
	{
		Name:        "Replies (Specific Thread)",
		Description: "Notify when a reply is posted in this thread",
		Method:      "internal",
		Pattern:     "reply:/forum/topic/{topicid}/thread/{threadid}/*",
	},

	// Private Forum
	{
		Name:        "Private Topic Created",
		Description: "Notify when a new private topic is created",
		Method:      "internal",
		Pattern:     "private topic create:/private/*",
	},

	// News
	{
		Name:        "New News Post",
		Description: "Notify when a new news post is created",
		Method:      "internal",
		Pattern:     "new post:/news/*",
	},
	{
		Name:        "Reply to News",
		Description: "Notify when a reply is posted to a news item",
		Method:      "internal",
		Pattern:     "reply:/news/news/*",
	},

	// Linker
	{
		Name:        "New Link Added",
		Description: "Notify when a new link is added",
		Method:      "internal",
		Pattern:     "add:/linker/*",
	},
	{
		Name:        "Reply to Link",
		Description: "Notify when a reply is posted to a link",
		Method:      "internal",
		Pattern:     "reply:/linker/*",
	},

	// FAQ
	{
		Name:        "New FAQ Question",
		Description: "Notify when a new FAQ question is created",
		Method:      "internal",
		Pattern:     "create:/faq/*",
	},
	{
		Name:        "New FAQ Ask",
		Description: "Notify when a user asks a question",
		Method:      "internal",
		Pattern:     "ask:/faq/*",
	},

	// Admin
	{
		Name:        "Admin Notifications",
		Description: "Receive general admin notifications",
		Method:      "internal",
		Pattern:     "notify:/admin/*",
		IsAdmin:     true,
	},
	{
		Name:        "User Registration",
		Description: "Notify when a new user registers",
		Method:      "internal",
		Pattern:     "register:/auth/register",
		IsAdmin:     true,
	},
}
