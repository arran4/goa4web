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
		Actions []common.Action
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum - Topic %d Grants", tid)
	data := Data{CoreData: cd, TopicID: int32(tid), Actions: []common.Action{common.ActionSee, common.ActionView, common.ActionReply, common.ActionPost, common.ActionEdit}}
	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		log.Printf("ListGrants: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	for _, g := range grants {
		if common.Section(g.Section) == common.SectionForum && g.Item.Valid && common.Item(g.Item.String) == common.ItemTopic && g.ItemID.Valid && g.ItemID.Int32 == int32(tid) {
			gi := GrantInfo{Grant: g}
			if g.UserID.Valid {
				if u, err := queries.SystemGetUserByID(r.Context(), g.UserID.Int32); err == nil {
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
