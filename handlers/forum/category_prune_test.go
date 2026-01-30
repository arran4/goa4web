package forum

import (
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestCategoryTreePruneEmpty(t *testing.T) {
	cats := []*db.Forumcategory{
		{Idforumcategory: 1, ForumcategoryIdforumcategory: 0},
		{Idforumcategory: 2, ForumcategoryIdforumcategory: 1},
		{Idforumcategory: 3, ForumcategoryIdforumcategory: 1},
		{Idforumcategory: 4, ForumcategoryIdforumcategory: 0},
	}
	topics := []*ForumtopicPlus{
		{Idforumtopic: 10, ForumcategoryIdforumcategory: 3},
	}
	ct := NewCategoryTree(cats, topics)
	ct.PruneEmpty()
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
	if len(root1.Categories) != 1 || root1.Categories[0].Idforumcategory != 3 {
		t.Fatalf("unexpected children for root 1: %#v", root1.Categories)
	}
}
