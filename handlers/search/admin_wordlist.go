package search

import (
	"database/sql"
	_ "embed"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

type WordCount struct {
	Word  sql.NullString
	Count int32
}

func adminSearchWordListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Rows       []WordCount
		Letters    []string
		CurrentLtr string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{}
	cd.PageTitle = "Search Word List"
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
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
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
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
		prefRows, err2 := queries.AdminWordListWithCountsByPrefix(r.Context(), db.AdminWordListWithCountsByPrefixParams{
			Prefix: letter,
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err2 != nil {
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
		for _, r := range prefRows {
			rows = append(rows, WordCount{Word: r.Word, Count: r.Count})
		}
	} else {
		totalCount, err = queries.AdminCountWordList(r.Context())
		if err != nil {
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
		allRows, err2 := queries.AdminWordListWithCounts(r.Context(), db.AdminWordListWithCountsParams{
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err2 != nil {
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
		for _, r := range allRows {
			rows = append(rows, WordCount{Word: r.Word, Count: r.Count})
		}
	}
	data.Rows = rows

	base := "/admin/search/list"
	vals := url.Values{}
	if letter != "" {
		vals.Set("letter", letter)
	}

	pagBase := base
	if len(vals) > 0 {
		pagBase += "?" + vals.Encode()
	}

	cd.Pagination = &common.PageNumberPagination{
		TotalItems:  int(totalCount),
		PageSize:    pageSize,
		CurrentPage: page,
		BaseURL:     pagBase,
		ParamName:   "page",
	}

	AdminSearchWordListPageTmpl.Handle(w, r, data)
}

const AdminSearchWordListPageTmpl tasks.Template = "admin/searchWordListPage.gohtml"

// adminSearchWordListDownloadPage sends the full word list as a text file.
func adminSearchWordListDownloadPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	rows, err := queries.AdminCompleteWordList(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
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
