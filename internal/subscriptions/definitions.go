package subscriptions

import (
	"regexp"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

type Definition struct {
	Name        string
	Description string
	Pattern     string
	IsAdminOnly bool
}

// SubscriptionInstance represents a concrete subscription (e.g. to Topic #1).
type SubscriptionInstance struct {
	Parameters map[string]string // e.g. "topicid" -> "1"
	Methods    []string          // e.g. ["internal", "email"]
	Original   string            // Original DB pattern string
}

// HasMethod checks if the instance has the given method.
func (si *SubscriptionInstance) HasMethod(method string) bool {
	for _, m := range si.Methods {
		if m == method {
			return true
		}
	}
	return false
}

// SubscriptionGroup groups instances by their definition.
type SubscriptionGroup struct {
	*Definition
	Instances []*SubscriptionInstance
}

var Definitions = []Definition{
	// Forum
	{
		Name:        "New Threads (All)",
		Description: "Notify when a new thread is created in any topic",
		Pattern:     "create thread:/forum/topic/*",
	},
	{
		Name:        "New Threads (Specific Topic)",
		Description: "Notify when a new thread is created in this topic",
		Pattern:     "create thread:/forum/topic/{topicid}/*",
	},
	{
		Name:        "Replies (All)",
		Description: "Notify when a reply is posted in any thread",
		Pattern:     "reply:/forum/topic/*/thread/*",
	},
	{
		Name:        "Replies (Specific Thread)",
		Description: "Notify when a reply is posted in this thread",
		Pattern:     "reply:/forum/topic/{topicid}/thread/{threadid}/*",
	},

	// Private Forum
	{
		Name:        "Private Topic Created",
		Description: "Notify when a new private topic is created",
		Pattern:     "private topic create:/private/*",
	},

	// News
	{
		Name:        "New News Post",
		Description: "Notify when a new news post is created",
		Pattern:     "new post:/news/*",
	},
	{
		Name:        "Reply to News",
		Description: "Notify when a reply is posted to a news item",
		Pattern:     "reply:/news/news/*",
	},

	// Linker
	{
		Name:        "New Link Added",
		Description: "Notify when a new link is added",
		Pattern:     "add:/linker/*",
	},
	{
		Name:        "Reply to Link",
		Description: "Notify when a reply is posted to a link",
		Pattern:     "reply:/linker/*",
	},

	// FAQ
	{
		Name:        "New FAQ Question",
		Description: "Notify when a new FAQ question is created",
		Pattern:     "create:/faq/*",
	},
	{
		Name:        "New FAQ Ask",
		Description: "Notify when a user asks a question",
		Pattern:     "ask:/faq/*",
	},

	// Admin
	{
		Name:        "Admin Notifications",
		Description: "Receive general admin notifications",
		Pattern:     "notify:/admin/*",
		IsAdminOnly: true,
	},
	{
		Name:        "User Registration",
		Description: "Notify when a new user registers",
		Pattern:     "register:/auth/register",
		IsAdminOnly: true,
	},
}

// GetUserSubscriptions groups user subscriptions.
// dbSubs is a list of rows from the subscriptions table.
func GetUserSubscriptions(dbSubs []*db.ListSubscriptionsByUserRow) []*SubscriptionGroup {
	groups := make(map[string]*SubscriptionGroup)

	// Initialize groups for all definitions
	for i := range Definitions {
		def := &Definitions[i]
		groups[def.Pattern] = &SubscriptionGroup{
			Definition: def,
			Instances:  []*SubscriptionInstance{},
		}
	}

	for _, sub := range dbSubs {
		def, params := MatchDefinition(sub.Pattern)
		if def == nil {
			// Handle unknown/custom patterns? For now, skip or log.
			continue
		}

		group := groups[def.Pattern]

		// Find existing instance with same parameters
		var instance *SubscriptionInstance
		for _, inst := range group.Instances {
			if equalParams(inst.Parameters, params) {
				instance = inst
				break
			}
		}

		if instance == nil {
			instance = &SubscriptionInstance{
				Parameters: params,
				Methods:    []string{},
				Original:   sub.Pattern,
			}
			group.Instances = append(group.Instances, instance)
		}

		// Add method if not present
		found := false
		for _, m := range instance.Methods {
			if m == sub.Method {
				found = true
				break
			}
		}
		if !found {
			instance.Methods = append(instance.Methods, sub.Method)
		}
	}

	// Convert map to slice, preserving order of Definitions
	var result []*SubscriptionGroup
	for i := range Definitions {
		result = append(result, groups[Definitions[i].Pattern])
	}
	return result
}

// MatchDefinition attempts to match a pattern string against known definitions.
// Returns the definition and extracted parameters (if any).
func MatchDefinition(pattern string) (*Definition, map[string]string) {
	for i := range Definitions {
		def := &Definitions[i]
		if params, ok := matchPattern(def.Pattern, pattern); ok {
			return def, params
		}
	}
	return nil, nil
}

// matchPattern checks if 'pattern' matches 'template' (which may contain {param}).
// It treats '*' in template as a literal wildcard character if the pattern has it too,
// OR as a matching wildcard if the template has it and the pattern has content.
//
// Current simple implementation:
// Template: create thread:/forum/topic/{topicid}/*
// Pattern:  create thread:/forum/topic/123/*
func matchPattern(template, pattern string) (map[string]string, bool) {
	// Convert template to regex
	// Escape special chars
	regexStr := regexp.QuoteMeta(template)

	// Replace \{param\} with named capturing group
	// We need to un-escape the braces we just escaped
	regexStr = strings.ReplaceAll(regexStr, "\\{", "{")
	regexStr = strings.ReplaceAll(regexStr, "\\}", "}")

	// Replace {name} with (?P<name>[^/]+)
	paramRegex := regexp.MustCompile(`\{([a-zA-Z0-9]+)\}`)
	regexStr = paramRegex.ReplaceAllString(regexStr, `(?P<$1>[^/]+)`)

	// Handle standard wildcard *
	// If template has *, it matches anything remaining or specific segment?
	// The definitions use * at the end mostly.
	// Let's replace \* (escaped) with .*
	regexStr = strings.ReplaceAll(regexStr, "\\*", ".*")

	regexStr = "^" + regexStr + "$"
	re := regexp.MustCompile(regexStr)

	match := re.FindStringSubmatch(pattern)
	if match == nil {
		return nil, false
	}

	params := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i > 0 && name != "" {
			params[name] = match[i]
		}
	}
	return params, true
}

func equalParams(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
