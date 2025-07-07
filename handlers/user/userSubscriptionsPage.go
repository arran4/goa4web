package user

import (
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	db "github.com/arran4/goa4web/internal/db"
)

type subscriptionOption struct {
	Name    string
	Pattern string
	Path    string
	Task    string
}

var userSubscriptionOptions = []subscriptionOption{
	{Name: "Blog posts", Pattern: "post:/blog/*", Path: "blogs", Task: common.TaskSubscribeBlogs},
	{Name: "Writings", Pattern: "post:/writing/*", Path: "writings", Task: common.TaskSubscribeWritings},
	{Name: "News posts", Pattern: "post:/news/*", Path: "news", Task: common.TaskSubscribeNews},
	{Name: "Image board posts", Pattern: "post:/image/*", Path: "images", Task: common.TaskSubscribeImages},
}

func userSubscriptionsPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.ListSubscriptionsByUser(r.Context(), uid)
	if err != nil {
		log.Printf("list subs: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		Subs    []*db.ListSubscriptionsByUserRow
		Options []subscriptionOption
		Error   string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Subs:     rows,
		Options:  userSubscriptionOptions,
		Error:    r.URL.Query().Get("error"),
	}
	if err := templates.RenderTemplate(w, "subscriptions.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func addSubscription(w http.ResponseWriter, r *http.Request, pattern string) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/subscriptions?error="+err.Error(), http.StatusSeeOther)
		return
	}
	method := r.PostFormValue("method")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := queries.InsertSubscription(r.Context(), db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: pattern, Method: method}); err != nil {
		log.Printf("insert sub: %v", err)
		http.Redirect(w, r, "/usr/subscriptions?error="+err.Error(), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/usr/subscriptions", http.StatusSeeOther)
}

func userSubscriptionsAddBlogsAction(w http.ResponseWriter, r *http.Request) {
	addSubscription(w, r, "post:/blog/*")
}

func userSubscriptionsAddWritingsAction(w http.ResponseWriter, r *http.Request) {
	addSubscription(w, r, "post:/writing/*")
}

func userSubscriptionsAddNewsAction(w http.ResponseWriter, r *http.Request) {
	addSubscription(w, r, "post:/news/*")
}

func userSubscriptionsAddImagesAction(w http.ResponseWriter, r *http.Request) {
	addSubscription(w, r, "post:/image/*")
}

func userSubscriptionsDeleteAction(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/subscriptions?error="+err.Error(), http.StatusSeeOther)
		return
	}
	idStr := r.PostFormValue("id")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if idStr == "" {
		http.Redirect(w, r, "/usr/subscriptions?error=missing id", http.StatusSeeOther)
		return
	}
	id, _ := strconv.Atoi(idStr)
	if err := queries.DeleteSubscriptionByID(r.Context(), db.DeleteSubscriptionByIDParams{UsersIdusers: uid, ID: int32(id)}); err != nil {
		log.Printf("delete sub: %v", err)
		http.Redirect(w, r, "/usr/subscriptions?error="+err.Error(), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/usr/subscriptions", http.StatusSeeOther)
}
