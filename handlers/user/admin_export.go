package user

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// gdprExportNote is included in exports to emphasise that the data is
// personal and must be handled carefully.
const gdprExportNote = "# Personal data export - handle according to GDPR"

// adminUsersExportPage streams all data for a single user in a zip archive for
// admins. The user ID is provided via the "uid" query parameter.
func adminUsersExportPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	uid, err := strconv.Atoi(r.URL.Query().Get("uid"))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	urow, err := queries.GetUserById(r.Context(), int32(uid))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	user := &db.User{Idusers: urow.Idusers, Email: urow.Email, Passwd: urow.Passwd, PasswdAlgorithm: urow.PasswdAlgorithm, Username: urow.Username}

	pref, _ := queries.GetPreferenceByUserID(r.Context(), int32(uid))
	langs, _ := queries.GetUserLanguages(r.Context(), int32(uid))
	perms, _ := queries.GetPermissionsByUserID(r.Context(), int32(uid))

	data := struct {
		Note        string           `json:"note"`
		User        *db.User         `json:"user"`
		Preference  *db.Preference   `json:"preference,omitempty"`
		Languages   []*db.Userlang   `json:"languages,omitempty"`
		Permissions []*db.Permission `json:"permissions,omitempty"`
	}{
		Note:        gdprExportNote,
		User:        user,
		Preference:  pref,
		Languages:   langs,
		Permissions: perms,
	}

	cats, _ := queries.FetchAllCategories(r.Context())
	catMap := make(map[int32]string)
	for _, c := range cats {
		catMap[c.Idwritingcategory] = c.Title.String
	}

	writings, _ := queries.GetAllWritingsByUser(r.Context(), int32(uid))
	type writingExport struct {
		*db.GetAllWritingsByUserRow
		Category string `json:"category"`
	}
	var ws []writingExport
	for _, wrow := range writings {
		ws = append(ws, writingExport{wrow, catMap[wrow.WritingcategoryIdwritingcategory]})
	}

	blogs, _ := queries.GetAllBlogEntriesByUser(r.Context(), int32(uid))
	threads, _ := queries.GetThreadsStartedByUser(r.Context(), int32(uid))
	comments, _ := queries.GetAllCommentsByUser(r.Context(), int32(uid))

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=user_%d.zip", uid))
	zw := zip.NewWriter(w)
	defer zw.Close()

	if f, err := zw.Create("user.json"); err == nil {
		_ = json.NewEncoder(f).Encode(data)
	}
	if f, err := zw.Create("writings.json"); err == nil {
		_ = json.NewEncoder(f).Encode(ws)
	}
	for _, wrow := range writings {
		if wrow.Writting.Valid {
			if f, err := zw.Create(fmt.Sprintf("writings/%d.html", wrow.Idwriting)); err == nil {
				_, _ = f.Write([]byte(wrow.Writting.String))
			}
		}
	}
	if f, err := zw.Create("blogs.json"); err == nil {
		_ = json.NewEncoder(f).Encode(blogs)
	}
	for _, b := range blogs {
		if b.Blog.Valid {
			if f, err := zw.Create(fmt.Sprintf("blogs/%d.html", b.Idblogs)); err == nil {
				_, _ = f.Write([]byte(b.Blog.String))
			}
		}
	}
	if f, err := zw.Create("threads.json"); err == nil {
		_ = json.NewEncoder(f).Encode(threads)
	}
	if f, err := zw.Create("comments.json"); err == nil {
		_ = json.NewEncoder(f).Encode(comments)
	}
}
