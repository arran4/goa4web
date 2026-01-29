package role_templates

import "sort"

// TemplateDef describes a named template of roles and grants.
type TemplateDef struct {
	Name        string
	Description string
	Roles       []RoleDef
}

// RoleDef describes a role definition within a template.
type RoleDef struct {
	Name        string
	CanLogin    bool
	IsAdmin     bool
	Description string
	Grants      []GrantDef
}

// GrantDef describes a single grant definition in a role template.
type GrantDef struct {
	Section string
	Item    string // can be empty
	Action  string
	ItemID  int32 // 0 for global/any
}

// Templates lists the available role templates by name.
var Templates = map[string]TemplateDef{
	"default": {
		Name:        "default",
		Description: "Basic setup with minimal news access.",
		Roles: []RoleDef{
			{
				Name:        "guest",
				CanLogin:    false,
				IsAdmin:     false,
				Description: "Read-only access to news.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					{Section: "faq", Item: "", Action: "search"},
					{Section: "faq", Item: "question/answer", Action: "see"},
				},
			},
			{
				Name:        "user",
				CanLogin:    true,
				IsAdmin:     false,
				Description: "Basic user.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					{Section: "faq", Item: "", Action: "search"},
					{Section: "faq", Item: "question/answer", Action: "see"},
				},
			},
			{
				Name:        "admin",
				CanLogin:    true,
				IsAdmin:     true,
				Description: "Administrator with news management rights.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "post"},
					{Section: "news", Item: "post", Action: "edit"},
					{Section: "news", Item: "post", Action: "reply"},
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					{Section: "faq", Item: "", Action: "search"},
					{Section: "faq", Item: "question/answer", Action: "see"},
				},
			},
		},
	},
	"simple-community": {
		Name:        "simple-community",
		Description: "Community setup with news, private forums, and labelling.",
		Roles: []RoleDef{
			{
				Name:        "guest",
				CanLogin:    false,
				IsAdmin:     false,
				Description: "Read-only access to public sections.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					{Section: "faq", Item: "", Action: "search"},
					{Section: "faq", Item: "question/answer", Action: "see"},
				},
			},
			{
				Name:        "user",
				CanLogin:    true,
				IsAdmin:     false,
				Description: "Standard user with access to private forums and labelling.",
				Grants: []GrantDef{
					// News reader
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					// Labeller
					{Section: "news", Item: "post", Action: "label"},
					{Section: "privateforum", Item: "topic", Action: "label"},
					// Private forum user
					{Section: "privateforum", Item: "topic", Action: "see"},
					{Section: "privateforum", Item: "topic", Action: "view"},
					// FAQ reader
					{Section: "faq", Item: "", Action: "search"},
					{Section: "faq", Item: "question/answer", Action: "see"},
				},
			},
			{
				Name:        "admin",
				CanLogin:    true,
				IsAdmin:     true,
				Description: "Administrator with full access and content management rights.",
				Grants: []GrantDef{
					// News writer
					{Section: "news", Item: "post", Action: "post"},
					{Section: "news", Item: "post", Action: "edit"},
					{Section: "news", Item: "post", Action: "reply"},
					// News reader
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					// Labeller
					{Section: "news", Item: "post", Action: "label"},
					{Section: "privateforum", Item: "topic", Action: "label"},
					// Private forum user
					{Section: "privateforum", Item: "topic", Action: "see"},
					{Section: "privateforum", Item: "topic", Action: "view"},
					{Section: "privateforum", Item: "topic", Action: "post"},
					{Section: "privateforum", Item: "topic", Action: "reply"},
					{Section: "privateforum", Item: "topic", Action: "edit"},
					// FAQ reader
					{Section: "faq", Item: "", Action: "search"},
					{Section: "faq", Item: "question/answer", Action: "see"},
				},
			},
			{
				Name:        "image-uploader",
				CanLogin:    false,
				IsAdmin:     false,
				Description: "Can upload images.",
				Grants: []GrantDef{
					{Section: "images", Item: "upload", Action: "post"},
				},
			},
		},
	},
	"news-only": {
		Name:        "news-only",
		Description: "Setup focused solely on news publishing and reading.",
		Roles: []RoleDef{
			{
				Name:        "guest",
				CanLogin:    false,
				IsAdmin:     false,
				Description: "News reader.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
				},
			},
			{
				Name:        "editor",
				CanLogin:    true,
				IsAdmin:     false,
				Description: "News content creator.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "post"},
					{Section: "news", Item: "post", Action: "edit"},
					{Section: "news", Item: "post", Action: "reply"},
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
				},
			},
			{
				Name:        "admin",
				CanLogin:    true,
				IsAdmin:     true,
				Description: "Administrator.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "post"},
					{Section: "news", Item: "post", Action: "edit"},
					{Section: "news", Item: "post", Action: "reply"},
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
				},
			},
		},
	},
	"read-only": {
		Name:        "read-only",
		Description: "Restrictive setup where almost everyone is a reader.",
		Roles: []RoleDef{
			{
				Name:        "guest",
				CanLogin:    false,
				IsAdmin:     false,
				Description: "Global reader.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					{Section: "privateforum", Item: "topic", Action: "see"},
					{Section: "privateforum", Item: "topic", Action: "view"},
				},
			},
			{
				Name:        "admin",
				CanLogin:    true,
				IsAdmin:     true,
				Description: "Administrator.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
					{Section: "privateforum", Item: "topic", Action: "see"},
					{Section: "privateforum", Item: "topic", Action: "view"},
				},
			},
		},
	},
}

// SortedTemplateNames returns the template keys in sorted order.
func SortedTemplateNames() []string {
	names := make([]string, 0, len(Templates))
	for name := range Templates {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
