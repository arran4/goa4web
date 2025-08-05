package common

// CanSearch reports whether cd permits running a search in the given section.
// It checks both a section specific grant and a global search grant.
func CanSearch(cd *CoreData, section Section) bool {
	if cd == nil {
		return false
	}
	if cd.HasGrant(section, ItemNone, ActionSearch, 0) {
		return true
	}
	return cd.HasGrant(SectionSearch, ItemNone, ActionSearch, 0)
}
