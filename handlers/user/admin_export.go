package user

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

// gdprExportNote is included in exports to emphasise that the data is
// personal and must be handled carefully.
const gdprExportNote = "# Personal data export - handle according to GDPR"

// adminUsersExportPage streams all data for a single user in a zip archive for
// admins. The user ID is provided via the "uid" query parameter.
func adminUsersExportPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	uid, err := strconv.Atoi(r.URL.Query().Get("uid"))
	if err != nil {
		log.Printf("parse uid: %v", err)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}

	row, err := queries.SystemGetUserByID(r.Context(), int32(uid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		log.Printf("current user: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	user := &db.User{
		Idusers:                row.Idusers,
		Username:               row.Username,
		PublicProfileEnabledAt: row.PublicProfileEnabledAt,
	}

	pref, err := queries.GetPreferenceForLister(r.Context(), int32(uid))
	if err != nil {
		log.Printf("load preference: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	langs, err := queries.GetUserLanguages(r.Context(), int32(uid))
	if err != nil {
		log.Printf("load languages: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	perms, err := queries.GetPermissionsByUserID(r.Context(), int32(uid))
	if err != nil {
		log.Printf("load permissions: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	// NOTE: Intentionally not including email in db.User struct as it wasn't there before
	// and db.User struct doesn't have an email field.
	// db.User definition:
	// type User struct {
	// 	Idusers                int32
	// 	Username               sql.NullString
	// 	DeletedAt              sql.NullTime
	// 	PublicProfileEnabledAt sql.NullTime
	// }
	data := struct {
		Note        string                          `json:"note"`
		User        *db.User                        `json:"user"`
		Preference  *db.Preference                  `json:"preference,omitempty"`
		Languages   []*db.UserLanguage              `json:"languages,omitempty"`
		Permissions []*db.GetPermissionsByUserIDRow `json:"permissions,omitempty"`
	}{
		Note:        gdprExportNote,
		User:        user,
		Preference:  pref,
		Languages:   langs,
		Permissions: perms,
	}

	// Previously using cd.WritingCategories() which used SystemListWritingCategories
	// without any user filtering. Replicating that behavior here.
	cats, err := queries.SystemListWritingCategories(r.Context(), db.SystemListWritingCategoriesParams{
		Limit:  math.MaxInt32,
		Offset: 0,
	})
	if err != nil {
		log.Printf("fetch categories: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	catMap := make(map[int32]string)
	for _, c := range cats {
		catMap[c.Idwritingcategory] = c.Title.String
	}

	writings, err := queries.AdminGetAllWritingsByAuthor(r.Context(), int32(uid))
	if err != nil {
		log.Printf("fetch writings: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	type writingExport struct {
		*db.AdminGetAllWritingsByAuthorRow
		Category string `json:"category"`
	}
	var ws []writingExport
	for _, wrow := range writings {
		ws = append(ws, writingExport{wrow, catMap[wrow.WritingCategoryID]})
	}

	blogs, err := queries.AdminGetAllBlogEntriesByUser(r.Context(), int32(uid))
	if err != nil {
		log.Printf("fetch blogs: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	threads, err := queries.AdminGetThreadsStartedByUser(r.Context(), int32(uid))
	if err != nil {
		log.Printf("fetch threads: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	comments, err := queries.AdminGetAllCommentsByUser(r.Context(), int32(uid))
	if err != nil {
		log.Printf("fetch comments: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=user_%d.zip", uid))
	zw := zip.NewWriter(w)
	defer zw.Close()

	if f, err := zw.Create("user.json"); err == nil {
		if err := json.NewEncoder(f).Encode(data); err != nil {
			log.Printf("write user.json: %v", err)
		}
	} else {
		log.Printf("create user.json: %v", err)
	}
	if f, err := zw.Create("writings.json"); err == nil {
		if err := json.NewEncoder(f).Encode(ws); err != nil {
			log.Printf("write writings.json: %v", err)
		}
	} else {
		log.Printf("create writings.json: %v", err)
	}
	for _, wrow := range writings {
		if wrow.Writing.Valid {
			if f, err := zw.Create(fmt.Sprintf("writings/%d.html", wrow.Idwriting)); err == nil {
				if _, err := f.Write([]byte(wrow.Writing.String)); err != nil {
					log.Printf("write writing %d: %v", wrow.Idwriting, err)
				}
			} else {
				log.Printf("create writing %d: %v", wrow.Idwriting, err)
			}
		}
	}
	if f, err := zw.Create("blogs.json"); err == nil {
		if err := json.NewEncoder(f).Encode(blogs); err != nil {
			log.Printf("write blogs.json: %v", err)
		}
	} else {
		log.Printf("create blogs.json: %v", err)
	}
	for _, b := range blogs {
		if b.Blog.Valid {
			if f, err := zw.Create(fmt.Sprintf("blogs/%d.html", b.Idblogs)); err == nil {
				if _, err := f.Write([]byte(b.Blog.String)); err != nil {
					log.Printf("write blog %d: %v", b.Idblogs, err)
				}
			} else {
				log.Printf("create blog %d: %v", b.Idblogs, err)
			}
		}
	}
	if f, err := zw.Create("threads.json"); err == nil {
		if err := json.NewEncoder(f).Encode(threads); err != nil {
			log.Printf("write threads.json: %v", err)
		}
	} else {
		log.Printf("create threads.json: %v", err)
	}
	if f, err := zw.Create("comments.json"); err == nil {
		if err := json.NewEncoder(f).Encode(comments); err != nil {
			log.Printf("write comments.json: %v", err)
		}
	} else {
		log.Printf("create comments.json: %v", err)
	}
}
