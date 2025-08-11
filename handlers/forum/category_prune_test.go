package forum

import (
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestCategoryTreePruneEmpty(t *testing.T) {
	cats := []*db.Forumcategory{
		{ID: 1, ParentCategoryID: 0},
		{ID: 2, ParentCategoryID: 1},
		{ID: 3, ParentCategoryID: 1},
		{ID: 4, ParentCategoryID: 0},
	}
	topics := []*ForumtopicPlus{
		{ID: 10, CategoryID: 3},
	}
	ct := NewCategoryTree(cats, topics)
	if _, ok := ct.CategoryLookup[4]; ok {
		t.Fatalf("category 4 should be pruned")
	}
	if _, ok := ct.CategoryLookup[2]; ok {
		t.Fatalf("category 2 should be pruned")
	}
	root1, ok := ct.CategoryLookup[1]
	if !ok {
		t.Fatalf("root category 1 missing")
	}
	if len(root1.Categories) != 1 || root1.Categories[0].ID != 3 {
		t.Fatalf("unexpected children for root 1: %#v", root1.Categories)
	}
}
