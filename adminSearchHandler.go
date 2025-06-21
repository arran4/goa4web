package main

import (
	"fmt"
	"net/http"
)

func adminSearchPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	renderTemplate(w, r, "adminSearchPage.gohtml", data)
}

/*
void addToGeneralSearch(a4webcont &cont, char *text, int fid, char *dbtable, char *forgienkey)
{
	if (fid == 0) return;
	a4string query("INSERT INTO %s (%s, searchwordlist_idsearchwordlist) VALUES ", dbtable, forgienkey);
	a4hashtable words;
	a4hashtable nowords;
	//nowords.set("this", (void*)1);
	breakupTextToWords(cont, text, words, nowords);
	char **keys = words.keys();
	int count = 0;
	while (*keys != NULL)
	{
		int wordid = getSearchWordByWordLowercased(cont, *keys);
		if (wordid == 0)
		{
			wordid = createSearchWord(cont, *keys);
		}
		if (wordid)
			query.pushf("%s(%d, %d)", count++ ? "," : "", fid, wordid);
		keys++;
	}
	if (count)
	{
		a4mysqlResult *result = cont.sql.query(query.raw());
		delete result;
	}
}

static void remakeSearchs(a4webcont &cont, char *idname, char *textbodyname, char *sourcetable, char* searchtable, char *forgienkey)
{
	a4string query("SELECT %s, %s FROM %s", idname, textbodyname, sourcetable);
	a4mysqlResult *result = cont.sql.query(query.raw());
	a4LinkedList<struct storeText *> queue;
	if (result->hasRow())
		do
		{
			struct storeText *tmp = (struct storeText *)malloc(sizeof(struct storeText));
			tmp->aid = atoiornull(result->getColumn(0));
			tmp->text = strdup(result->getColumn(1));
			queue.push(tmp);
		} while (result->nextRow());
	delete result;
	query.set("DELETE FROM %s", searchtable);
	result = cont.sql.query(query.raw());
	delete result;
	while (queue.total())
	{
		struct storeText *tmp = NULL;
		tmp = queue.shift();
		addToGeneralSearch(cont, tmp->text, tmp->aid, searchtable, forgienkey);
		free(tmp->text);
		free(tmp);
	}
}

static void remakeCommentsSearch(a4webcont &cont)
{
	remakeSearchs(cont, "idcomments", "text", "comments", "commentsSearch", "comments_idcomments");
}

static void remakeNewsSearch(a4webcont &cont)
{
	remakeSearchs(cont, "idsiteNews", "news", "siteNews", "siteNewsSearch", "siteNews_idsiteNews");
}

static void remakeBlogSearch(a4webcont &cont)
{
	remakeSearchs(cont, "idblogs", "blog", "blogs", "blogsSearch", "blogs_idblogs");
}

static void remakeWritingSearch(a4webcont &cont)
{
	remakeSearchs(cont, "idwriting", "concat(title, \" \", abstract, \" \", writting)", "writing", "writingSearch", "writing_idwriting");
}

static void remakeLinkerSearch(a4webcont &cont)
{
	remakeSearchs(cont, "idlinker", "concat(title, \" \", description)", "linker", "linkerSearch", "linker_idlinker");
}

*/

func adminSearchRemakeCommentsSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteCommentsSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteCommentsSearch: %w", err).Error())
	}
	if err := queries.RemakeCommentsSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeCommentsSearchInsert: %w", err).Error())
	}

	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
}
func adminSearchRemakeNewsSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteSiteNewsSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteSiteNewsSearch: %w", err).Error())
	}
	if err := queries.RemakeNewsSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeNewsSearchInsert: %w", err).Error())
	}

	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
}
func adminSearchRemakeBlogSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteBlogsSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteBlogsSearch: %w", err).Error())
	}
	if err := queries.RemakeBlogsSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeBlogsSearchInsert: %w", err).Error())
	}

	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
}
func adminSearchRemakeLinkerSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteLinkerSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteLinkerSearch: %w", err).Error())
	}
	if err := queries.RemakeLinkerSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeLinkerSearchInsert: %w", err).Error())
	}

	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
}
func adminSearchRemakeWritingSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteWritingSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteWritingSearch: %w", err).Error())
	}
	if err := queries.RemakeWritingSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeWritingSearchInsert: %w", err).Error())
	}

	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
}
