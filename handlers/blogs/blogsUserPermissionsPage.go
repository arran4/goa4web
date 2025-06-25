package blogs

import (
	"database/sql"
	"errors"
	"fmt"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/templates"
)

func GetPermissionsByUserIdAndSectionBlogsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(common.KeyCoreData).(*corecommon.CoreData)
	if !(cd.HasRole("writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	type Data struct {
		*corecommon.CoreData
		Rows   []*GetPermissionsByUserIdAndSectionBlogsRow
		Filter string
	}

	data := Data{
		CoreData: cd,
		Filter:   r.URL.Query().Get("level"),
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	rows, err := queries.GetPermissionsByUserIdAndSectionBlogs(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if data.Filter != "" {
		filtered := rows[:0]
		for _, row := range rows {
			if row.Level.String == data.Filter {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}

	data.Rows = rows

	CustomBlogIndex(data.CoreData, r)
	err = templates.RenderTemplate(w, "userPermissionsPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func UsersPermissionsPermissionUserAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	where := "blogs"
	level := r.PostFormValue("level")
	data := struct {
		*corecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		Back:     "/blogs/bloggers",
	}
	if u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername: %w", err).Error())
	} else if err := queries.PermissionUserAllow(r.Context(), PermissionUserAllowParams{
		UsersIdusers: u.Idusers,
		Section: sql.NullString{
			String: where,
			Valid:  true,
		},
		Level: sql.NullString{
			String: level,
			Valid:  true,
		},
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	}

	CustomBlogIndex(data.CoreData, r)

	err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func UsersPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*corecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		Back:     "/blogs/bloggers",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.PermissionUserDisallow(r.Context(), int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	CustomBlogIndex(data.CoreData, r)
	err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func UsersPermissionsBulkAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	names := strings.FieldsFunc(r.PostFormValue("usernames"), func(r rune) bool { return r == ',' || r == '\n' || r == ' ' || r == '\t' })
	level := r.PostFormValue("level")
	data := struct {
		*corecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		Back:     "/blogs/bloggers",
	}

	for _, n := range names {
		if n == "" {
			continue
		}
		u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: n})
		if err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername %s: %w", n, err).Error())
			continue
		}
		if err := queries.PermissionUserAllow(r.Context(), PermissionUserAllowParams{
			UsersIdusers: u.Idusers,
			Section:      sql.NullString{String: "blogs", Valid: true},
			Level:        sql.NullString{String: level, Valid: true},
		}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow %s: %w", n, err).Error())
		}
	}

	CustomBlogIndex(data.CoreData, r)
	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func UsersPermissionsBulkDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	permids := r.PostForm["permid"]
	data := struct {
		*corecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		Back:     "/blogs/bloggers",
	}

	for _, id := range permids {
		if id == "" {
			continue
		}
		permidi, err := strconv.Atoi(id)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi %s: %w", id, err).Error())
			continue
		}
		if err := queries.PermissionUserDisallow(r.Context(), int32(permidi)); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("permissionUserDisallow %s: %w", id, err).Error())
		}
	}

	CustomBlogIndex(data.CoreData, r)
	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
