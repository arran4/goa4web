package navigation

// RouterOptions defines an option that can apply changes to the navigation registry.
type RouterOptions interface {
	Apply(*Registry)
}

// IndexLinkOption represents an index link registration.
type IndexLinkOption struct {
	Name   string
	URL    string
	Weight int
}

func (o *IndexLinkOption) Apply(r *Registry) {
	r.RegisterIndexLink(o.Name, o.URL, o.Weight)
}

// NewIndexLink creates a new IndexLinkOption.
func NewIndexLink(name, url string, weight int) RouterOptions {
	return &IndexLinkOption{Name: name, URL: url, Weight: weight}
}

// IndexLinkWithViewPermissionOption represents an index link registration with view permission.
type IndexLinkWithViewPermissionOption struct {
	Name        string
	URL         string
	Weight      int
	ViewSection string
	ViewItem    string
}

func (o *IndexLinkWithViewPermissionOption) Apply(r *Registry) {
	r.RegisterIndexLinkWithViewPermission(o.Name, o.URL, o.Weight, o.ViewSection, o.ViewItem)
}

// NewIndexLinkWithViewPermission creates a new IndexLinkWithViewPermissionOption.
func NewIndexLinkWithViewPermission(name, url string, weight int, section, item string) RouterOptions {
	return &IndexLinkWithViewPermissionOption{Name: name, URL: url, Weight: weight, ViewSection: section, ViewItem: item}
}

// AdminControlCenterLinkOption represents an admin control center link registration.
type AdminControlCenterLinkOption struct {
	Section any
	Name    string
	URL     string
	Weight  int
}

func (o *AdminControlCenterLinkOption) Apply(r *Registry) {
	r.RegisterAdminControlCenter(o.Section, o.Name, o.URL, o.Weight)
}

// NewAdminControlCenterLink creates a new AdminControlCenterLinkOption.
func NewAdminControlCenterLink(section any, name, url string, weight int) RouterOptions {
	return &AdminControlCenterLinkOption{Section: section, Name: name, URL: url, Weight: weight}
}
