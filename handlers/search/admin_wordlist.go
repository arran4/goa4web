package search

import (
	"database/sql"
	_ "embed"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

type PageLink struct {
	Num    int
	Link   string
	Active bool
}

type WordCount struct {
	Word  sql.NullString
	Count int32
}

func adminSearchWordListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows       []WordCount
		NextLink   string
		PrevLink   string
		PageLinks  []PageLink
		Letters    []string
		CurrentLtr string
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	data.CoreData.PageTitle = "Search Word List"

	letters := make([]string, len(handlers.Alphabet))
	for i, c := range handlers.Alphabet {
		letters[i] = strings.ToUpper(string(c))
	}
	data.Letters = letters

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	letter := strings.ToLower(r.URL.Query().Get("letter"))
	if len(letter) > 1 {
		letter = letter[:1]
	}
	data.CurrentLtr = strings.ToUpper(letter)

	const pageSize = 1000

	offset := (page - 1) * pageSize

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if r.URL.Query().Get("download") != "" {
		rows, err := queries.AdminCompleteWordList(r.Context())
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=wordlist.txt")
		for _, row := range rows {
			if row.Valid {
				if _, err := w.Write([]byte(row.String + "\n")); err != nil {
					log.Printf("write wordlist row: %v", err)
				}
			}
		}
		return
	}

	var (
		rows       []WordCount
		err        error
		totalCount int64
	)
	if letter != "" {
		totalCount, err = queries.AdminCountWordListByPrefix(r.Context(), letter)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		prefRows, err2 := queries.AdminWordListWithCountsByPrefix(r.Context(), db.AdminWordListWithCountsByPrefixParams{
			Prefix: letter,
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err2 != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		for _, r := range prefRows {
			rows = append(rows, WordCount{Word: r.Word, Count: r.Count})
		}
	} else {
		totalCount, err = queries.AdminCountWordList(r.Context())
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		allRows, err2 := queries.AdminWordListWithCounts(r.Context(), db.AdminWordListWithCountsParams{
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err2 != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		for _, r := range allRows {
			rows = append(rows, WordCount{Word: r.Word, Count: r.Count})
		}
	}
	data.Rows = rows

	numPages := int((totalCount + int64(pageSize-1)) / int64(pageSize))

	base := "/admin/search/list"
	vals := url.Values{}
	if letter != "" {
		vals.Set("letter", letter)
	}

	for i := 1; i <= numPages; i++ {
		vals.Set("page", strconv.Itoa(i))
		data.PageLinks = append(data.PageLinks, PageLink{Num: i, Link: base + "?" + vals.Encode(), Active: i == page})
	}
	if page < numPages {
		vals.Set("page", strconv.Itoa(page+1))
		data.NextLink = base + "?" + vals.Encode()
	}
	if page > 1 {
		vals.Set("page", strconv.Itoa(page-1))
		data.PrevLink = base + "?" + vals.Encode()
	}

	handlers.TemplateHandler(w, r, "searchWordListPage.gohtml", data)
}

// adminSearchWordListDownloadPage sends the full word list as a text file.
func adminSearchWordListDownloadPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	rows, err := queries.AdminCompleteWordList(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=wordlist.txt")
	for _, row := range rows {
		if row.Valid {
			if _, err := w.Write([]byte(row.String + "\n")); err != nil {
				log.Printf("write wordlist row: %v", err)
			}
		}
	}
}
