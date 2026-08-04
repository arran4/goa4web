package user

import (
	"archive/zip"
	"context"
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

type userExportData struct {
	Note        string                          `json:"note"`
	User        *db.User                        `json:"user"`
	Preference  *db.Preference                  `json:"preference,omitempty"`
	Languages   []*db.UserLanguage              `json:"languages,omitempty"`
	Permissions []*db.GetPermissionsByUserIDRow `json:"permissions,omitempty"`
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

// Previously using cd.WritingCategories() which used SystemListWritingCategories
// without any user filtering. Replicating that behavior here.

type writingExport struct {
	*db.AdminGetAllWritingsByAuthorRow
	Category string `json:"category"`
}

func writeJSONToZip(zw *zip.Writer, filename string, data interface{}) error {
	f, err := zw.Create(filename)
	if err != nil {
		return fmt.Errorf("create %s: %w", filename, err)
	}
	if err := json.NewEncoder(f).Encode(data); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}
	return nil
}

func writeTextToZip(zw *zip.Writer, filename string, content string) error {
	f, err := zw.Create(filename)
	if err != nil {
		return fmt.Errorf("create %s: %w", filename, err)
	}
	if _, err := f.Write([]byte(content)); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}
	return nil
}

func fetchUserExportData(ctx context.Context, queries db.Querier, uid int32) (*userExportData, error) {
	row, err := queries.SystemGetUserByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("system get user by id: %w", err)
	}

	user := &db.User{
		Idusers:                row.Idusers,
		Username:               row.Username,
		PublicProfileEnabledAt: row.PublicProfileEnabledAt,
	}

	pref, err := queries.GetPreferenceForLister(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get preference for lister: %w", err)
	}
	langs, err := queries.GetUserLanguages(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get user languages: %w", err)
	}
	perms, err := queries.GetPermissionsByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get permissions by user id: %w", err)
	}

	return &userExportData{
		Note:        gdprExportNote,
		User:        user,
		Preference:  pref,
		Languages:   langs,
		Permissions: perms,
	}, nil
}

func fetchWritingsExportData(ctx context.Context, queries db.Querier, uid int32) ([]writingExport, error) {
	cats, err := queries.SystemListWritingCategories(ctx, db.SystemListWritingCategoriesParams{
		Limit:  math.MaxInt32,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("system list writing categories: %w", err)
	}
	catMap := make(map[int32]string)
	for _, c := range cats {
		catMap[c.Idwritingcategory] = c.Title.String
	}

	writings, err := queries.AdminGetAllWritingsByAuthor(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("admin get all writings by author: %w", err)
	}

	var ws []writingExport
	for _, wrow := range writings {
		ws = append(ws, writingExport{wrow, catMap[wrow.WritingCategoryID]})
	}
	return ws, nil
}

func streamExportZip(w http.ResponseWriter, uid int32, data *userExportData, ws []writingExport, blogs []*db.AdminGetAllBlogEntriesByUserRow, threads []*db.Forumthread, comments []*db.AdminGetAllCommentsByUserRow) {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=user_%d.zip", uid))
	zw := zip.NewWriter(w)
	defer zw.Close()

	if err := writeJSONToZip(zw, "user.json", data); err != nil {
		log.Printf("%v", err)
	}
	if err := writeJSONToZip(zw, "writings.json", ws); err != nil {
		log.Printf("%v", err)
	}
	for _, wrow := range ws {
		if wrow.Writing.Valid {
			if err := writeTextToZip(zw, fmt.Sprintf("writings/%d.html", wrow.Idwriting), wrow.Writing.String); err != nil {
				log.Printf("%v", err)
			}
		}
	}
	if err := writeJSONToZip(zw, "blogs.json", blogs); err != nil {
		log.Printf("%v", err)
	}
	for _, b := range blogs {
		if b.Blog.Valid {
			if err := writeTextToZip(zw, fmt.Sprintf("blogs/%d.html", b.Idblogs), b.Blog.String); err != nil {
				log.Printf("%v", err)
			}
		}
	}
	if err := writeJSONToZip(zw, "threads.json", threads); err != nil {
		log.Printf("%v", err)
	}
	if err := writeJSONToZip(zw, "comments.json", comments); err != nil {
		log.Printf("%v", err)
	}
}

func adminUsersExportPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := ctx.Value(consts.KeyCoreData).(*common.CoreData).Queries()

	uid, err := strconv.Atoi(r.URL.Query().Get("uid"))
	if err != nil {
		log.Printf("parse uid: %v", err)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}

	data, err := fetchUserExportData(ctx, queries, int32(uid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		log.Printf("fetch user export data: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	ws, err := fetchWritingsExportData(ctx, queries, int32(uid))
	if err != nil {
		log.Printf("fetch writings export data: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	blogs, err := queries.AdminGetAllBlogEntriesByUser(ctx, int32(uid))
	if err != nil {
		log.Printf("fetch blogs: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	threads, err := queries.AdminGetThreadsStartedByUser(ctx, int32(uid))
	if err != nil {
		log.Printf("fetch threads: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	comments, err := queries.AdminGetAllCommentsByUser(ctx, int32(uid))
	if err != nil {
		log.Printf("fetch comments: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	streamExportZip(w, int32(uid), data, ws, blogs, threads, comments)
}
