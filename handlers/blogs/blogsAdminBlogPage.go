package blogs

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminBlogPage shows details for a single blog entry including grants.
func AdminBlogPage(w http.ResponseWriter, r *http.Request) {
	type GrantInfo struct {
		*db.Grant
		Username sql.NullString
		RoleName sql.NullString
	}
	type Data struct {
		Blog   *db.GetBlogEntryForListerByIDRow
		Grants []GrantInfo
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	blogID, err := strconv.Atoi(vars["blog"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	queries := cd.Queries()
	blog, err := queries.GetBlogEntryForListerByID(r.Context(), db.GetBlogEntryForListerByIDParams{
		ListerID: cd.UserID,
		ID:       int32(blogID),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Blog not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Blog %d Admin", blog.Idblogs)
	data := Data{Blog: blog}
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		log.Printf("ListGrants: %v", err)
	} else {
		var roles []*db.Role
		if roles, err = cd.AllRoles(); err != nil {
			log.Printf("AllRoles: %v", err)
		}
		for _, g := range grants {
			if g.Section == "blogs" && g.Item.Valid && g.Item.String == "entry" && g.ItemID.Valid && g.ItemID.Int32 == blog.Idblogs {
				gi := GrantInfo{Grant: g}
				if g.UserID.Valid {
					if u, err := queries.SystemGetUserByID(r.Context(), g.UserID.Int32); err == nil {
						gi.Username = sql.NullString{String: u.Username.String, Valid: true}
					}
				}
				if g.RoleID.Valid && roles != nil {
					for _, role := range roles {
						if role.ID == g.RoleID.Int32 {
							gi.RoleName = sql.NullString{String: role.Name, Valid: true}
							break
						}
					}
				}
				data.Grants = append(data.Grants, gi)
			}
		}
	}
	handlers.TemplateHandler(w, r, "blogsAdminBlogPage.gohtml", data)
}
