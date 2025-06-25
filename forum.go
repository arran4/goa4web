package goa4web

import (
	"bytes"
	"database/sql"
	"fmt"

	"golang.org/x/exp/slices"
)

func processCommentFullQuote(username, text string) string {
	var out bytes.Buffer
	var quote bytes.Buffer
	var it, bc, nlc int

	for it < len(text) {
		switch text[it] {
		case ']':
			bc--
		case '[':
			bc++
		case '\\':
			if it+1 < len(text) {
				if text[it+1] == '[' || text[it+1] == ']' {
					out.WriteByte(text[it+1])
					it++
				}
			}
		case '\n':
			if bc == 0 && nlc == 1 {
				quote.WriteString(processCommentQuote(username, out.String()))
				out.Reset()
			}
			nlc++
			it++
			continue
		case '\r':
			it++
			continue
		case ' ':
			fallthrough
		default:
			if nlc != 0 {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
				nlc = 0
			}
			out.WriteByte(text[it])
		}
		it++
	}
	quote.WriteString(processCommentQuote(username, out.String()))
	return quote.String()
}

func processCommentQuote(username string, text string) string {
	return fmt.Sprintf("[quoteof \"%s\" %s]\n", username, text)
}

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
