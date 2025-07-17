package user

import (
	"log"
	"net/http"
	"strconv"

	handlers "github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type subscriptionOption struct {
	Name    string
	Pattern string
	Path    string
}

var userSubscriptionOptions = []subscriptionOption{
	{Name: "New blog posts", Pattern: "post:/blog/*", Path: "blogs"},
	{Name: "New articles", Pattern: "post:/writing/*", Path: "writings"},
	{Name: "New news posts", Pattern: "post:/news/*", Path: "news"},
	{Name: "New image board posts", Pattern: "post:/image/*", Path: "images"},
}

type SubscriptionsUpdateTask struct{ tasks.TaskString }
type SubscriptionsDeleteTask struct{ tasks.TaskString }

var (
	subscriptionsUpdateTask = &SubscriptionsUpdateTask{TaskString: TaskUpdate}
	subscriptionsDeleteTask = &SubscriptionsDeleteTask{TaskString: TaskDelete}
)

func userSubscriptionsPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	rows, err := queries.ListSubscriptionsByUser(r.Context(), uid)
	if err != nil {
		log.Printf("list subs: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	subMap := make(map[string]bool)
	for _, s := range rows {
		key := s.Pattern + "|" + s.Method
		subMap[key] = true
	}
	data := struct {
		*handlers.CoreData
		Subs    []*db.ListSubscriptionsByUserRow
		Options []subscriptionOption
		SubMap  map[string]bool
		Error   string
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
		Subs:     rows,
		Options:  userSubscriptionOptions,
		SubMap:   subMap,
		Error:    r.URL.Query().Get("error"),
	}
	handlers.TemplateHandler(w, r, "subscriptions.gohtml", data)
}

func (SubscriptionsUpdateTask) Action(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/subscriptions?error="+err.Error(), http.StatusSeeOther)
		return
	}
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	existing, err := queries.ListSubscriptionsByUser(r.Context(), uid)
	if err != nil {
		log.Printf("list subs: %v", err)
		http.Redirect(w, r, "/usr/subscriptions?error="+err.Error(), http.StatusSeeOther)
		return
	}
	have := make(map[string]bool)
	for _, s := range existing {
		have[s.Pattern+"|"+s.Method] = true
	}
	methods := []string{"internal", "email"}
	for _, opt := range userSubscriptionOptions {
		for _, m := range methods {
			key := opt.Path + "_" + m
			want := r.PostFormValue(key) != ""
			hkey := opt.Pattern + "|" + m
			if want && !have[hkey] {
				if err := queries.InsertSubscription(r.Context(), db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: opt.Pattern, Method: m}); err != nil {
					log.Printf("insert sub: %v", err)
					http.Redirect(w, r, "/usr/subscriptions?error="+err.Error(), http.StatusSeeOther)
					return
				}
			} else if !want && have[hkey] {
				if err := queries.DeleteSubscription(r.Context(), db.DeleteSubscriptionParams{UsersIdusers: uid, Pattern: opt.Pattern, Method: m}); err != nil {
					log.Printf("delete sub: %v", err)
					http.Redirect(w, r, "/usr/subscriptions?error="+err.Error(), http.StatusSeeOther)
					return
				}
			}
		}
	}
	http.Redirect(w, r, "/usr/subscriptions", http.StatusSeeOther)
}

func (SubscriptionsDeleteTask) Action(w http.ResponseWriter, r *http.Request) {
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
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
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
