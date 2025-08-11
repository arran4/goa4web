package forum

import (
	"database/sql"
	"github.com/arran4/goa4web/internal/db"
	"golang.org/x/exp/slices"
)

type ForumtopicPlus struct {
	ID           int32
	LastAuthorID int32
	CategoryID   int32
	Title        sql.NullString
	Description  sql.NullString
	// DisplayTitle optionally overrides Title for display purposes.
	DisplayTitle       string
	Threads            sql.NullInt32
	Comments           sql.NullInt32
	Lastaddition       sql.NullTime
	LastAuthorUsername sql.NullString
	Edit               bool
}

type ForumcategoryPlus struct {
	*db.Forumcategory
	Categories []*ForumcategoryPlus
	Topics     []*ForumtopicPlus
	Edit       bool
}

type CategoryTree struct {
	CategoryChildrenLookup map[int32][]*ForumcategoryPlus
	CategoryLookup         map[int32]*ForumcategoryPlus
}

func NewCategoryTree(categoryRows []*db.Forumcategory, topicRows []*ForumtopicPlus) *CategoryTree {
	categoryTree := new(CategoryTree)
	categoryTree.CategoryChildrenLookup = map[int32][]*ForumcategoryPlus{}
	categoryTree.CategoryLookup = map[int32]*ForumcategoryPlus{}
	for _, row := range categoryRows {
		fcp := &ForumcategoryPlus{
			Forumcategory: row,
			Categories:    nil,
			Edit:          false,
		}
		categoryTree.CategoryChildrenLookup[row.ParentCategoryID] = append(categoryTree.CategoryChildrenLookup[row.ParentCategoryID], fcp)
		categoryTree.CategoryLookup[row.ID] = fcp
	}
	for _, row := range topicRows {
		c, ok := categoryTree.CategoryLookup[row.CategoryID]
		if !ok || c == nil {
			continue
		}
		c.Topics = append(c.Topics, row)
	}
	for parentId, children := range categoryTree.CategoryChildrenLookup {
		parent, ok := categoryTree.CategoryLookup[parentId]
		if !ok {
			continue
		}
		parent.Categories = children
	}
	categoryTree.PruneEmpty()
	return categoryTree
}

func (ct *CategoryTree) pruneCategory(cat *ForumcategoryPlus) bool {
	keep := len(cat.Topics) > 0
	var filtered []*ForumcategoryPlus
	for _, c := range cat.Categories {
		if ct.pruneCategory(c) {
			filtered = append(filtered, c)
			keep = true
		} else {
			delete(ct.CategoryLookup, c.ID)
			delete(ct.CategoryChildrenLookup, c.ID)
		}
	}
	cat.Categories = filtered
	if !keep {
		delete(ct.CategoryLookup, cat.ID)
		delete(ct.CategoryChildrenLookup, cat.ID)
	}
	return keep
}

// PruneEmpty removes categories that contain no visible topics and no
// subcategories with visible topics.
func (ct *CategoryTree) PruneEmpty() {
	roots := ct.CategoryChildrenLookup[0]
	var filtered []*ForumcategoryPlus
	for _, root := range roots {
		if ct.pruneCategory(root) {
			filtered = append(filtered, root)
		} else {
			delete(ct.CategoryLookup, root.ID)
			delete(ct.CategoryChildrenLookup, root.ID)
		}
	}
	ct.CategoryChildrenLookup[0] = filtered
}

func (ct *CategoryTree) CategoryRoots(categoryId int32) (result []*ForumcategoryPlus) {
	catId := categoryId
	for {
		cat, ok := ct.CategoryLookup[catId]
		if !ok {
			break
		}
		result = append(result, cat)
		catId = cat.ParentCategoryID
		if catId == 0 {
			break
		}
	}
	slices.Reverse(result)
	return
}
