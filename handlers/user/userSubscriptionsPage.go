package user

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/internal/db"
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

func userSubscriptionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session, ok := cd.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := cd.Queries()
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
		*common.CoreData
		Subs    []*db.ListSubscriptionsByUserRow
		Options []subscriptionOption
		SubMap  map[string]bool
		Error   string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Subs:     rows,
		Options:  userSubscriptionOptions,
		SubMap:   subMap,
		Error:    r.URL.Query().Get("error"),
	}
	handlers.TemplateHandler(w, r, "subscriptions.gohtml", data)
}
