package common

// CanSearch reports whether cd permits running a search in the given section.
// It checks both a section specific grant and a global search grant.
func CanSearch(cd *CoreData, section string) bool {
	if cd == nil {
		return false
	}
	if cd.HasGrant(section, "", "search", 0) {
		return true
	}
	return cd.HasGrant("search", "", "search", 0)
}

// CanSearch reports whether the current CoreData instance permits running a search in the given section.
// It forwards to the package-level CanSearch helper for reuse in templates.
func (cd *CoreData) CanSearch(section string) bool {
	return CanSearch(cd, section)
}
