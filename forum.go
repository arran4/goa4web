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
	CategoryChildrenLookup map[int32][]*ForumcategoryPlus
	CategoryLookup         map[int32]*ForumcategoryPlus
	//TopicLookup         map[int32]*ForumtopicPlus
}

func NewCategoryTree(categoryRows []*Forumcategory, topicRows []*showTableTopicsRow) *CategoryTree {
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
		tp := &ForumtopicPlus{
			showTableTopicsRow: row,
		}
		c, ok := categoryTree.CategoryLookup[row.ForumcategoryIdforumcategory]
		if !ok || c == nil {
			continue
		}
		c.Topics = append(c.Topics, tp)
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
	/*
		categoryBreadcrumbs.tmpl
		static void printCategoryRoots(a4webcont &cont, int categoryno)
		{
			a4string query("SELECT c3.idforumcategory, c3.title, c2.idforumcategory, c2.title, c1.title FROM forumcategory c1 "
					"LEFT JOIN forumcategory c2 ON c2.idforumcategory=c1.forumcategory_idforumcategory "
					"LEFT JOIN forumcategory c3 ON c3.idforumcategory=c2.forumcategory_idforumcategory "
					"WHERE c1.idforumcategory=\"%d\"", categoryno);
			a4mysqlResult *result = cont.sql.query(query.raw(), false);
			printf("[");
			printf("<a href=\"?\">Forum</a>:");
			if (atoiornull(result->getColumn(0)))
			{
				a4code2html tmp;
				tmp.input.set("%s", result->getColumn(1));
				tmp.process();
				printf("<a href=\"?\">%s</a>:", tmp.output.raw());
			}
			if (atoiornull(result->getColumn(2)))
			{
				a4code2html tmp;
				tmp.input.set("%s", result->getColumn(3));
				tmp.process();
				printf("<a href=\"?\">%s</a>:", tmp.output.raw());
			}
			a4code2html tmp;
			tmp.input.set("%s", result->getColumn(4));
			tmp.process();
			printf("<a href=\"?\">(%s/Refresh)</a>", tmp.output.raw());
			printf("]<br>\n");
			delete result;
		}
	*/
	cat, ok := ct.CategoryLookup[categoryId]
	if !ok {
		return []*ForumcategoryPlus{}
	}
	for cat.Idforumcategory != 0 {
		cat, ok := ct.CategoryLookup[cat.ForumcategoryIdforumcategory]
		if !ok {
			break
		}
		result = append(result, cat)
	}
	return
}

func (ct *CategoryTree) TopicRoots(topicId int32) []*ForumcategoryPlus {
	/*
		categoryBreadcrumbs.tmpl
		static void printTopicRoots(a4webcont &cont, int topicno)
		{
			a4string query("SELECT c3.idforumcategory, c3.title, c2.idforumcategory, c2.title, c1.idforumcategory, c1.title, t.title FROM forumtopic t "
					"LEFT JOIN forumcategory c1 ON c1.idforumcategory=t.forumcategory_idforumcategory "
					"LEFT JOIN forumcategory c2 ON c2.idforumcategory=c1.forumcategory_idforumcategory "
					"LEFT JOIN forumcategory c3 ON c3.idforumcategory=c2.forumcategory_idforumcategory "
					"WHERE t.idforumtopic=\"%d\"", topicno);
			a4mysqlResult *result = cont.sql.query(query.raw(), false);
			printf("[");
			printf("<a href=\"?\">Forum</a>:");
			if (atoiornull(result->getColumn(0)))
			{
				a4code2html tmp;
				tmp.input.set("%s", result->getColumn(1));
				tmp.process();
				printf("<a href=\"?\">%s</a>:", tmp.output.raw());
			}
			if (atoiornull(result->getColumn(2)))
			{
				a4code2html tmp;
				tmp.input.set("%s", result->getColumn(3));
				tmp.process();
				printf("<a href=\"?\">%s</a>:", tmp.output.raw());
			}
			if (atoiornull(result->getColumn(4)))
			{
				a4code2html tmp;
				tmp.input.set("%s", result->getColumn(5));
				tmp.process();
				printf("<a href=\"?\">%s</a>:", tmp.output.raw());
			}
			a4code2html tmp;
			tmp.input.set("%s", result->getColumn(6));
			tmp.process();
			printf("<a href=\"?topic=%d\">(%s/Refresh)</a>", topicno, tmp.output.raw());
			printf("]<br>\n");
			delete result;
		}
		static void printThreadRoots(a4webcont &cont, int threadno)
		{
			a4string query("SELECT c3.idforumcategory, c3.title, c2.idforumcategory, c2.title, c1.idforumcategory, c1.title, t.idforumtopic, t.title FROM forumthread th "
					"LEFT JOIN forumtopic t ON idforumtopic=th.forumtopic_idforumtopic "
					"LEFT JOIN forumcategory c1 ON c1.idforumcategory=t.forumcategory_idforumcategory "
					"LEFT JOIN forumcategory c2 ON c2.idforumcategory=c1.forumcategory_idforumcategory "
					"LEFT JOIN forumcategory c3 ON c3.idforumcategory=c2.forumcategory_idforumcategory "
					"WHERE th.idforumthread=\"%d\"", threadno);
			a4mysqlResult *result = cont.sql.query(query.raw(), false);
			printf("[");
			printf("<a href=\"?\">Forum</a>:");
			if (atoiornull(result->getColumn(0)))
			{
				a4code2html tmp;
				tmp.input.set("%s", result->getColumn(1));
				tmp.process();
				printf("<a href=\"?\">%s</a>:", tmp.output.raw());
			}
			if (atoiornull(result->getColumn(2)))
			{
				a4code2html tmp;
				tmp.input.set("%s", result->getColumn(3));
				tmp.process();
				printf("<a href=\"?\">%s</a>:", tmp.output.raw());
			}
			if (atoiornull(result->getColumn(4)))
			{
				a4code2html tmp;
				tmp.input.set("%s", result->getColumn(5));
				tmp.process();
				printf("<a href=\"?\">%s</a>:", tmp.output.raw());
			}
			if (atoiornull(result->getColumn(6)))
			{
				a4code2html tmp;
				tmp.input.set("%s", result->getColumn(7));
				tmp.process();
				printf("<a href=\"?topic=%s\">%s</a>:", result->getColumn(6), tmp.output.raw());
			}
			printf("<a href=\"?thread=%d\">(This thread/Refresh)</a>", threadno);
			printf("]<br>\n");
			printf("%s\n", cont.sql.error());
			delete result;
		}

	*/
	panic("not implemented")
}
