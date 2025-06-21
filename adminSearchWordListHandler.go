package main

import (
	_ "embed"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"log"
	"net/http"
	"strconv"
	"strings"
)

func adminSearchWordListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows     []*WordListWithCountsRow
		NextLink string
		PrevLink string
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if r.URL.Query().Get("download") != "" {
		rows, err := queries.CompleteWordList(r.Context())
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=wordlist.txt")
		for _, row := range rows {
			if row.Valid {
				_, _ = w.Write([]byte(row.String + "\n"))
			}
		}
		return
	}
	const pageSize = 1000
	rows, err := queries.WordListWithCounts(r.Context(), WordListWithCountsParams{
		Limit:  pageSize + 1,
		Offset: int32(offset),
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	data.Rows = rows

	base := "/admin/search/list"
	if hasMore {
		if strings.Contains(base, "?") {
			data.NextLink = base + "&offset=" + strconv.Itoa(offset+pageSize)
		} else {
			data.NextLink = base + "?offset=" + strconv.Itoa(offset+pageSize)
		}
	}
	if offset > 0 {
		if strings.Contains(base, "?") {
			data.PrevLink = base + "&offset=" + strconv.Itoa(offset-pageSize)
		} else {
			data.PrevLink = base + "?offset=" + strconv.Itoa(offset-pageSize)
		}
	}

	err = renderTemplate(w, r, "adminSearchWordListPage.gohtml", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// adminSearchWordListDownloadPage sends the full word list as a text file.
func adminSearchWordListDownloadPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.CompleteWordList(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=wordlist.txt")
	for _, row := range rows {
		if row.Valid {
			_, _ = w.Write([]byte(row.String + "\n"))
		}
	}
}
