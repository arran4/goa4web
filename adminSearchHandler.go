package main

import (
	"log"
	"net/http"
)

func adminSearchPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if err := compiledTemplates.ExecuteTemplate(w, "adminSearchPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

/*
void breakupTextToWords(a4webcont &cont, char*text, a4hashtable &words, a4hashtable &nowords)
{
	a4code2html decoder;
	decoder.codeType = ct_wordsonly;
	decoder.input.set("%s", text);
	decoder.Process();
	a4string word;
	decoder.output.itteratorGotoStart();
	int loop = 1;
	int c;
	for (c = decoder.output.itteratorGet(); loop; c = decoder.output.itteratorGetNext())
	{
		if (c == EOF)
		{
			loop = 0;
		}
		if (isalnum(c))
		{
			word.pushf("%c", tolower(c));
		} else {
			if (word.length() > 2)
				if (nowords.get(word.raw()) == NULL)
					words.set(word.raw(), (void*)1);
			word.clear();
		}
	}
}

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
		int wordid = getWordID(cont, *keys);
		if (wordid == 0)
		{
			wordid = addWord(cont, *keys);
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
	//TODO
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
func adminSearchRemakeNewsSearchPage(w http.ResponseWriter, r *http.Request) {
	//TODO
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
func adminSearchRemakeBlogSearchPage(w http.ResponseWriter, r *http.Request) {
	//TODO
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
func adminSearchRemakeLinkerSearchPage(w http.ResponseWriter, r *http.Request) {
	//TODO
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
func adminSearchRemakeWritingSearchPage(w http.ResponseWriter, r *http.Request) {
	//TODO
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
