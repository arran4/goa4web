package forum

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminTopicsPage shows all forum topics for management.
func AdminTopicsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Topics"
	queries := cd.Queries()
	offset := cd.Offset()
	ps := cd.PageSize()
	total, err := queries.AdminCountForumTopics(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	rows, err := queries.AdminListForumTopics(r.Context(), db.AdminListForumTopicsParams{Limit: int32(ps), Offset: int32(offset)})
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	numPages := int((total + int64(ps) - 1) / int64(ps))
	currentPage := offset/ps + 1
	base := "/admin/forum/topics"
	for i := 1; i <= numPages; i++ {
		cd.PageLinks = append(cd.PageLinks, common.PageLink{Num: i, Link: fmt.Sprintf("%s?offset=%d", base, (i-1)*ps), Active: i == currentPage})
	}
	if offset+ps < int(total) {
		cd.NextLink = fmt.Sprintf("%s?offset=%d", base, offset+ps)
	}
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("%s?offset=%d", base, offset-ps)
		cd.StartLink = base + "?offset=0"
	}

	data := struct {
		Topics []*db.Forumtopic
	}{
		Topics: rows,
	}

	handlers.TemplateHandler(w, r, "forum/adminTopicsPage.gohtml", data)
}

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
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
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
	handlers.TemplateHandler(w, r, "forum/adminTopicDeletePage.gohtml", data)
}

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
