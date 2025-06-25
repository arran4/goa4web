package goa4web

import (
	"database/sql"
	"golang.org/x/exp/slices"
)

type ForumtopicPlus struct {
	Idforumtopic                 int32
	Lastposter                   int32
	ForumcategoryIdforumcategory int32
	Title                        sql.NullString
	Description                  sql.NullString
	Threads                      sql.NullInt32
	Comments                     sql.NullInt32
	Lastaddition                 sql.NullTime
	Lastposterusername           sql.NullString
	Seelevel                     sql.NullInt32
	Level                        sql.NullInt32
	Edit                         bool
}

type ForumcategoryPlus struct {
	*Forumcategory
	Categories []*ForumcategoryPlus
	Topics     []*ForumtopicPlus
	Edit       bool
}

type CategoryTree struct {
	//Categories           []*ForumcategoryPlus
	//Topics               []*ForumtopicPlus
	CategoryChildrenLookup map[int32][]*ForumcategoryPlus
	CategoryLookup         map[int32]*ForumcategoryPlus
	//TopicLookup         map[int32]*ForumtopicPlus
}

func NewCategoryTree(categoryRows []*Forumcategory, topicRows []*ForumtopicPlus) *CategoryTree {
	categoryTree := new(CategoryTree)
	categoryTree.CategoryChildrenLookup = map[int32][]*ForumcategoryPlus{}
	categoryTree.CategoryLookup = map[int32]*ForumcategoryPlus{}
	for _, row := range categoryRows {
		fcp := &ForumcategoryPlus{
			Forumcategory: row,
			Categories:    nil,
			Edit:          false,
		}
		categoryTree.CategoryChildrenLookup[row.ForumcategoryIdforumcategory] = append(categoryTree.CategoryChildrenLookup[row.ForumcategoryIdforumcategory], fcp)
		categoryTree.CategoryLookup[row.Idforumcategory] = fcp
	}
	for _, row := range topicRows {
		c, ok := categoryTree.CategoryLookup[row.ForumcategoryIdforumcategory]
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
	return categoryTree
}

func (ct *CategoryTree) CategoryRoots(categoryId int32) (result []*ForumcategoryPlus) {
	catId := categoryId
	for {
		cat, ok := ct.CategoryLookup[catId]
		if !ok {
			break
		}
		result = append(result, cat)
		catId = cat.ForumcategoryIdforumcategory
		if catId == 0 {
			break
		}
	}
	slices.Reverse(result)
	return
}
