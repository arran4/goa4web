package forum

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func AdminTopicGrantsPage(w http.ResponseWriter, r *http.Request) {
	type GrantInfo struct {
		*db.Grant
		Username sql.NullString
		RoleName sql.NullString
	}
	type Data struct {
		*common.CoreData
		TopicID int32
		Grants  []GrantInfo
		Roles   []*db.Role
		Actions []string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum - Topic %d Grants", tid)
	data := Data{CoreData: cd, TopicID: int32(tid), Actions: []string{"see", "view", "reply", "post", "edit"}}
	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		log.Printf("ListGrants: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, g := range grants {
		if g.Section == "forum" && g.Item.Valid && g.Item.String == "topic" && g.ItemID.Valid && g.ItemID.Int32 == int32(tid) {
			gi := GrantInfo{Grant: g}
			if g.UserID.Valid {
				if u, err := queries.GetUserById(r.Context(), g.UserID.Int32); err == nil {
					gi.Username = sql.NullString{String: u.Username.String, Valid: true}
				}
			}
			if g.RoleID.Valid && data.Roles != nil {
				for _, r := range data.Roles {
					if r.ID == g.RoleID.Int32 {
						gi.RoleName = sql.NullString{String: r.Name, Valid: true}
						break
					}
				}
			}
			data.Grants = append(data.Grants, gi)
		}
	}
	handlers.TemplateHandler(w, r, "adminTopicGrantsPage.gohtml", data)
}
