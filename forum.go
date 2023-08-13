package main

import (
	"bytes"
	"fmt"
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
			switch text[it+1] {
			case '[', ']':
				out.WriteByte(text[it+1])
				it++
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
	*showTableTopicsRow
	Edit bool
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
	CategoryParentLookup map[int32][]*ForumcategoryPlus
	CategoryLookup       map[int32]*ForumcategoryPlus
}

func NewCategoryTree(categoryRows []*Forumcategory, topicRows []*showTableTopicsRow) *CategoryTree {
	categoryTree := new(CategoryTree)
	categoryTree.CategoryParentLookup = map[int32][]*ForumcategoryPlus{}
	categoryTree.CategoryLookup = map[int32]*ForumcategoryPlus{}
	for _, row := range categoryRows {
		fcp := &ForumcategoryPlus{
			Forumcategory: row,
			Categories:    nil,
			Edit:          false,
		}
		categoryTree.CategoryParentLookup[row.ForumcategoryIdforumcategory] = append(categoryTree.CategoryParentLookup[row.ForumcategoryIdforumcategory], fcp)
		categoryTree.CategoryLookup[row.Idforumcategory] = fcp
	}
	for _, row := range topicRows {
		tp := &ForumtopicPlus{
			showTableTopicsRow: row,
		}
		c, ok := categoryTree.CategoryLookup[row.ForumcategoryIdforumcategory]
		if !ok || c == nil {
			continue
		}
		c.Topics = append(c.Topics, tp)
	}
	for parentId, children := range categoryTree.CategoryParentLookup {
		parent, ok := categoryTree.CategoryLookup[parentId]
		if !ok {
			continue
		}
		parent.Categories = children
	}
	return categoryTree
}
