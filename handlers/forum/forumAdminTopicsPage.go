package forum

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

type AdminTopicDisplay struct {
	*db.Forumtopic

	IsPrivate    bool
	Participants string
	AccessInfo   string
	DisplayTitle string

	// Optional but handy for the list view / future tweaks
	CategoryLabel string
}

// AdminTopicsPage shows all forum topics for management.
func AdminTopicsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Topics"
	queries := cd.Queries()
	offset := cd.Offset()
	ps := cd.PageSize()

	total, err := queries.AdminCountForumTopics(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	rows, err := queries.AdminListForumTopics(r.Context(), db.AdminListForumTopicsParams{
		Limit:  int32(ps),
		Offset: int32(offset),
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	categories, err := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: 0})
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	categoryMap := make(map[int32]string, len(categories))
	for _, category := range categories {
		if category.Title.Valid {
			categoryMap[category.Idforumcategory] = category.Title.String
		}
	}

	// pagination links
	base := "/admin/forum/topics"
	cd.Pagination = &common.OffsetPagination{
		TotalItems: int(total),
		PageSize:   ps,
		Offset:     offset,
		BaseURL:    base,
	}

	topics := make([]*AdminTopicDisplay, 0, len(rows))
	for _, row := range rows {
		display := &AdminTopicDisplay{
			Forumtopic: row,
			IsPrivate:  row.Handler == "private",
		}

		// Title (don’t explode on NULL)
		if row.Title.Valid && strings.TrimSpace(row.Title.String) != "" {
			display.DisplayTitle = row.Title.String
		} else {
			display.DisplayTitle = "(untitled)"
		}

		// Category label (for non-private)
		if !display.IsPrivate {
			if label, ok := categoryMap[row.ForumcategoryIdforumcategory]; ok && strings.TrimSpace(label) != "" {
				display.CategoryLabel = label
			} else {
				display.CategoryLabel = "(none)"
			}
		}

		// Grants / access info
		grants, err := queries.AdminGetTopicGrants(r.Context(), sql.NullInt32{Int32: row.Idforumtopic, Valid: true})
		if err != nil {
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}

		if display.IsPrivate {
			// Private: show participants (use the grants list)
			var participants []string
			for _, g := range grants {
				if g.Username.Valid {
					participants = append(participants, g.Username.String)
				}
			}
			if len(participants) > 0 {
				display.Participants = strings.Join(participants, ", ")
			} else {
				display.Participants = "No participants"
			}
			// AccessInfo not needed for private topics
		} else {
			// Public: build an access summary from grants
			hasAnyone := false
			hasUsers := false
			var roles []string
			var users []string

			for _, g := range grants {
				if g.RoleName.Valid {
					rn := strings.ToLower(strings.TrimSpace(g.RoleName.String))
					switch rn {
					case "anyone":
						hasAnyone = true
					case "user", "users":
						hasUsers = true
					default:
						roles = append(roles, g.RoleName.String)
					}
				} else if g.Username.Valid {
					users = append(users, g.Username.String)
				}
			}

			if hasAnyone {
				display.AccessInfo = "Public (Anyone)"
			} else if hasUsers {
				display.AccessInfo = "Users"
			} else {
				parts := make([]string, 0, 2)
				if len(roles) > 0 {
					parts = append(parts, "Roles: "+strings.Join(roles, ", "))
				}
				if len(users) > 0 {
					parts = append(parts, "Users: "+strings.Join(users, ", "))
				}
				if len(parts) > 0 {
					display.AccessInfo = strings.Join(parts, "; ")
				} else if display.CategoryLabel != "" {
					// fallback that still gives the admin *some* signal
					display.AccessInfo = "Category: " + display.CategoryLabel
				} else {
					display.AccessInfo = "Default/Global"
				}
			}
		}

		topics = append(topics, display)
	}

	data := struct {
		Topics []*AdminTopicDisplay
	}{
		Topics: topics,
	}

	ForumAdminTopicsPageTmpl.Handle(w, r, data)
}

const ForumAdminTopicsPageTmpl tasks.Template = "forum/adminTopicsPage.gohtml"

func AdminTopicEditPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	languageID, _ := strconv.Atoi(r.PostFormValue("language"))

	if err := cd.Queries().AdminUpdateForumTopic(r.Context(), db.AdminUpdateForumTopicParams{
		Title:                        sql.NullString{String: name, Valid: true},
		Description:                  sql.NullString{String: desc, Valid: true},
		ForumcategoryIdforumcategory: int32(cid),
		TopicLanguageID:              sql.NullInt32{Int32: int32(languageID), Valid: languageID != 0},
		Idforumtopic:                 int32(tid),
	}); err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}

	http.Redirect(w, r, "/admin/forum/topics", http.StatusSeeOther)
}

func AdminTopicCreatePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	// derive section from base path, handling private forum mapping
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	section := strings.TrimPrefix(base, "/")
	if section == "private" {
		section = "privateforum"
	}
	allowed, err := UserCanCreateTopic(r.Context(), cd.Queries(), section, int32(pcid), uid)
	if err != nil {
		log.Printf("UserCanCreateTopic error: %v", err)
		w.WriteHeader(http.StatusForbidden)
		handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
		return
	}
	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
		return
	}
	languageID, _ := strconv.Atoi(r.PostFormValue("language"))
	topicID, err := cd.Queries().AdminCreateForumTopic(r.Context(), db.AdminCreateForumTopicParams{
		ForumcategoryID: int32(pcid),
		LanguageID:      sql.NullInt32{Int32: int32(languageID), Valid: languageID != 0},
		Title:           sql.NullString{String: name, Valid: true},
		Description:     sql.NullString{String: desc, Valid: true},
		Handler:         "",
	})
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	if topicID == 0 {
		w.WriteHeader(http.StatusForbidden)
		handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
		return
	}
	http.Redirect(w, r, "/admin/forum/topics", http.StatusSeeOther)
}

func AdminTopicDeleteConfirmPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "/admin/forum/topics", err)
		return
	}
	cd.PageTitle = "Confirm forum topic delete"
	data := struct {
		Message      string
		ConfirmLabel string
		Back         string
	}{
		Message:      "Are you sure you want to delete forum topic " + strconv.Itoa(tid) + "?",
		ConfirmLabel: "Confirm delete",
		Back:         "/admin/forum/topics/topic/" + strconv.Itoa(tid),
	}
	ForumAdminTopicDeletePageTmpl.Handle(w, r, data)
}

const ForumAdminTopicDeletePageTmpl tasks.Template = "forum/adminTopicDeletePage.gohtml"

func AdminTopicDeletePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	if r.FormValue("cascade") == "true" {
		if err := cd.Queries().DeleteThreadsByTopicID(r.Context(), int32(tid)); err != nil {
			handlers.RedirectSeeOtherWithError(w, r, "", err)
			return
		}
	}
	if err := cd.Queries().AdminDeleteForumTopic(r.Context(), int32(tid)); err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	http.Redirect(w, r, "/admin/forum/topics", http.StatusSeeOther)
}
