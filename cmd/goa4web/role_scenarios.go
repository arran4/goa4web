package main

type ScenarioDef struct {
	Name        string
	Description string
	Roles       []RoleDef
}

type RoleDef struct {
	Name        string
	CanLogin    bool
	IsAdmin     bool
	Description string
	Grants      []GrantDef
}

type GrantDef struct {
	Section string
	Item    string // can be empty
	Action  string
	ItemID  int32 // 0 for global/any
}

var roleScenarios = map[string]ScenarioDef{
	"default": {
		Name:        "default",
		Description: "Standard setup with guest, user, and admin roles.",
		Roles: []RoleDef{
			{
				Name:     "guest",
				CanLogin: false,
				IsAdmin:  false,
				Description: "Read-only access to public sections.",
				Grants: []GrantDef{
					{Section: "news", Item: "post", Action: "see"},
					{Section: "news", Item: "post", Action: "view"},
				},
			},
			{
				Name:     "user",
				CanLogin: true,
				IsAdmin:  false,
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
				},
			},
			{
				Name:     "admin",
				CanLogin: true,
				IsAdmin:  true,
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
				},
			},
		},
	},
}
