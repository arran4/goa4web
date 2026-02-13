package internal

import "strings"

// TODO move to internal or somewhere more appropriate

// Entry represents a single bookmark.
type Entry struct {
	Url  string
	Name string
}

// Category groups related bookmark entries under a name.
type Category struct {
	Name    string
	Entries []*Entry
}

// Column holds a set of bookmark categories.
type Column struct {
	Categories []*Category
}

// ParseColumns converts a raw bookmark list into structured columns.
// The list format is:
//
//	Category: <name>\n
//	<url> <title>\n
//	...
//
// Columns are separated by lines containing only "Column".
func ParseColumns(bookmarks string) []*Column {
	lines := strings.Split(bookmarks, "\n")
	result := []*Column{{}}
	var currentCategory *Category

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.EqualFold(line, "column") {
			result = append(result, &Column{})
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		if strings.EqualFold(parts[0], "Category:") {
			categoryName := strings.Join(parts[1:], " ")
			if currentCategory == nil {
				currentCategory = &Category{Name: categoryName}
			} else if currentCategory.Name != "" {
				result[len(result)-1].Categories = append(result[len(result)-1].Categories, currentCategory)
				currentCategory = &Category{Name: categoryName}
			} else {
				currentCategory.Name = categoryName
			}
		} else if currentCategory != nil {
			entry := &Entry{Url: parts[0], Name: parts[0]}
			if len(parts) > 1 {
				entry.Name = strings.Join(parts[1:], " ")
			}
			currentCategory.Entries = append(currentCategory.Entries, entry)
		}
	}

	if currentCategory != nil && currentCategory.Name != "" {
		result[len(result)-1].Categories = append(result[len(result)-1].Categories, currentCategory)
	}
	return result
}
