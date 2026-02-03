package subscriptions

import (
	"regexp"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

type Definition struct {
	Name                  string
	Description           string
	Pattern               string
	IsAdminOnly           bool
	Mandatory             bool
	HideIfNone            bool
	Upgrade               func(params map[string]string) string
	Legacy                bool
	SupportsAutoSubscribe bool
}

// Parameter represents a single parameter in a subscription pattern.
type Parameter struct {
	Key      string // e.g. "topicid"
	Value    string // e.g. "1"
	Resolved string // e.g. "General Discussion"
	Link     string // e.g. "/forum/topic/1"
}

// SubscriptionInstance represents a concrete subscription (e.g. to Topic #1).
type SubscriptionInstance struct {
	Parameters []Parameter // List of extracted parameters
	Methods    []string    // e.g. ["internal", "email"]
	Original   string      // Original DB pattern string
	UpgradeTo  string      // If set, this subscription can be upgraded to this pattern
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
	Name      string
	Instances []*SubscriptionInstance
}

var Definitions = []Definition{
	// Forum
	{
		Name:        "New Threads (Specific Topic)",
		Description: "Notify when a new thread is created in this topic",
		Pattern:     "create thread:/forum/topic/{topicid}*",
	},
	{
		Name:        "New Threads (All)",
		Description: "Notify when a new thread is created in any topic",
		Pattern:     "create thread:/forum/topic/*",
	},
	{
		Name:        "Replies (Specific Thread)",
		Description: "Notify when a reply is posted in this thread",
		Pattern:     "reply:/forum/topic/{topicid}/thread/{threadid}*",
	},
	{
		Name:        "Replies (All)",
		Description: "Notify when a reply is posted in any thread",
		Pattern:     "reply:/forum/topic/*/thread/*",
	},
	{
		Name:        "Edit Reply",
		Description: "Notify when a reply is edited",
		Pattern:     "edit reply:/forum/topic/*/thread*",
	},

	// Private Forum
	{
		Name:        "Private Topic Created",
		Description: "Notify when a new private topic is created",
		Pattern:     "private topic create:/private*",
	},
	{
		Name:                  "New Threads (Private Topic)",
		Description:           "Notify when a new thread is created in this private topic",
		Pattern:               "create thread:/private/topic/{topicid}*",
		SupportsAutoSubscribe: true,
	},
	{
		Name:        "Replies (Private Thread)",
		Description: "Notify when a reply is posted in this private thread",
		Pattern:     "reply:/private/topic/{topicid}/thread/{threadid}*",
	},

	// Blog
	{
		Name:        "New Blog Post",
		Description: "Notify when a new blog post is created",
		Pattern:     "post:/blog/*",
	},

	// Writing
	{
		Name:        "New Article",
		Description: "Notify when a new article is created",
		Pattern:     "post:/writing/*",
	},

	// Image Board
	{
		Name:        "New Image Post",
		Description: "Notify when a new image post is created",
		Pattern:     "post:/image/*",
	},

	// News
	{
		Name:        "New News Post",
		Description: "Notify when a new news post is created",
		Pattern:     "new post:/news/*",
	},
	{
		Name:        "New News Post (Legacy)",
		Description: "Notify when a new news post is created",
		Pattern:     "post:/news/*",
	},
	{
		Name:        "Reply to News",
		Description: "Notify when a reply is posted to a news item",
		Pattern:     "reply:/news/news/*",
	},
	{
		Name:        "Edit News Post",
		Description: "Notify when a news post is edited",
		Pattern:     "edit post:/news/news/*",
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

	// Legacy
	{
		Name:        "Write Reply",
		Description: "Notify when a reply is written (Legacy)",
		Pattern:     "write reply:/forum/topic/{topicid}/thread/{threadid}/*",
		HideIfNone:  true,
		Upgrade: func(params map[string]string) string {
			return "reply:/forum/topic/" + params["topicid"] + "/thread/" + params["threadid"] + "/*"
		},
		Legacy: true,
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
		Name:        "Admin: User Role Granted",
		Description: "Notify when a role is granted to a user",
		Pattern:     "user allow:/*",
		IsAdminOnly: true,
	},
	{
		Name:        "Admin: User Role Revoked",
		Description: "Notify when a role is revoked from a user",
		Pattern:     "user disallow:/*",
		IsAdminOnly: true,
	},
	{
		Name:        "Admin: Category Deleted",
		Description: "Notify when a category is deleted",
		Pattern:     "delete category:/*",
		IsAdminOnly: true,
	},
	{
		Name:        "Admin: New Image Board Post",
		Description: "Notify when a new image is uploaded",
		Pattern:     "upload image:/*",
		IsAdminOnly: true,
	},
	{
		Name:        "User Registration",
		Description: "Notify when a new user registers",
		Pattern:     "register:/auth/register",
		IsAdminOnly: true,
	},
	{
		Name:        "Password Reset",
		Description: "Notify when a user requests a password reset",
		Pattern:     "password reset:/auth/reset",
		IsAdminOnly: true,
		Mandatory:   true,
	},
	{
		Name:        "Email Verification",
		Description: "Notify when an email verification is requested",
		Pattern:     "email verification:/auth/verify_email",
		IsAdminOnly: true,
		Mandatory:   true,
	},
	{
		Name:        "User Approval",
		Description: "Notify when user approval is needed",
		Pattern:     "user approval:/admin/user_approval/*",
		IsAdminOnly: true,
	},
	{
		Name:        "Role Grant",
		Description: "Notify when a role is granted",
		Pattern:     "role grant:/admin/role_grant/*",
		IsAdminOnly: true,
	},
	{
		Name:        "Reports",
		Description: "Notify when content is reported",
		Pattern:     "report:/admin/report/*",
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
		name := def.Name
		if def.Legacy {
			name += " (Legacy)"
		}
		groups[def.Pattern] = &SubscriptionGroup{
			Definition: def,
			Name:       name,
			Instances:  []*SubscriptionInstance{},
		}
	}

	for _, sub := range dbSubs {
		def, params := MatchDefinition(sub.Pattern)
		// For unknown patterns, create a temporary definition group
		var group *SubscriptionGroup
		if def == nil {
			unknownKey := "unknown:" + sub.Pattern
			if _, exists := groups[unknownKey]; !exists {
				groups[unknownKey] = &SubscriptionGroup{
					Definition: &Definition{
						Name:    "Unknown: " + sub.Pattern,
						Pattern: sub.Pattern,
					},
					Name:      "Unknown: " + sub.Pattern,
					Instances: []*SubscriptionInstance{},
				}
			}
			group = groups[unknownKey]
			def = group.Definition
		} else {
			group = groups[def.Pattern]
		}

		// Find existing instance with same parameters
		var instance *SubscriptionInstance
		for _, inst := range group.Instances {
			if equalParams(inst.Parameters, params) {
				instance = inst
				break
			}
		}

		if instance == nil {
			var paramList []Parameter
			for k, v := range params {
				paramList = append(paramList, Parameter{Key: k, Value: v})
			}
			instance = &SubscriptionInstance{
				Parameters: paramList,
				Methods:    []string{},
				Original:   sub.Pattern,
			}
			if def.Upgrade != nil {
				instance.UpgradeTo = def.Upgrade(params)
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

	// Convert map to slice, preserving order of Definitions first
	var result []*SubscriptionGroup
	seen := make(map[string]bool)

	// Add predefined definitions
	for i := range Definitions {
		key := Definitions[i].Pattern
		if group, ok := groups[key]; ok {
			result = append(result, group)
			seen[key] = true
		}
	}

	// Add any unknown/custom ones that were found
	for key, group := range groups {
		if !seen[key] {
			result = append(result, group)
		}
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
func matchPattern(template, pattern string) (map[string]string, bool) {
	// Convert template to regex
	regexStr := regexp.QuoteMeta(template)

	// Replace \{param\} with named capturing group
	regexStr = strings.ReplaceAll(regexStr, "\\{", "{")
	regexStr = strings.ReplaceAll(regexStr, "\\}", "}")

	// Replace {name} with (?P<name>[^/*]+)
	paramRegex := regexp.MustCompile(`\{([a-zA-Z0-9]+)\}`)
	regexStr = paramRegex.ReplaceAllString(regexStr, `(?P<$1>[^/*]+)`)

	// Handle standard wildcard *
	// Replace \* with .*
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

func equalParams(a []Parameter, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, p := range a {
		if b[p.Key] != p.Value {
			return false
		}
	}
	return true
}
